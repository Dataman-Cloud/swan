package raft

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Dataman-Cloud/swan/log"
	"github.com/Dataman-Cloud/swan/store"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/etcdserver/stats"
	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/coreos/etcd/pkg/idutil"
	"github.com/coreos/etcd/pkg/types"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/rafthttp"
	"github.com/coreos/etcd/snap"
	"github.com/coreos/etcd/wal"
	"github.com/coreos/etcd/wal/walpb"
	events "github.com/docker/go-events"
	"github.com/docker/swarmkit/watch"
	"github.com/pivotal-golang/clock"
	"golang.org/x/net/context"
)

const (
	MaxTransactionBytes = 1.5 * 1024 * 1024
)

type LeadershipState int

const (
	// indicates that the node is a raft leader
	IsLeader LeadershipState = iota

	// indicates that the node is a raft follower
	IsFollower
)

var (
	// returns when an operation was submitted but the node was stopped in the meantime
	ErrStopped = errors.New("raft: failed to process the request: node is stopped")

	// returns when an operation was submitted but the node lost leader status before it became committed
	ErrLostLeadership = errors.New("raft: failed to process the request: node losts leader status")

	// returns when a raft internal message is too large to be send
	ErrRequestToolarge = errors.New("raft: raft message is too large and can't be send")

	// returns when the node is not yet part of a raft cluster
	ErrNoRaftMember = errors.New("raft: node is mnot yet part of a raft cluster")

	// returns when the cluster has no elected leader
	ErrNoClusterLeader = errors.New("raft: no elected cluster leader")
)

type Node struct {
	id        int      // client id for raft session
	peers     []string // raft peers URLS
	waldir    string   // path to WAL directory
	snapdir   string   // path to  sanpshot directory
	lastIndex uint64   // index fo log at start

	confState     raftpb.ConfState
	snapshotIndex uint64
	appliedIndex  uint64

	// raft backing config and raftnode
	raftNode    raft.Node
	Config      *raft.Config
	raftStorage *raft.MemoryStorage
	wal         *wal.WAL
	snapshotter *snap.Snapshotter
	snapCount   uint64
	transport   *rafthttp.Transport

	wait                *wait
	reqIDGen            *idutil.Generator
	signalledLeadership uint32
	isMember            uint32
	ticker              clock.Ticker
	leadershipBroadcast *watch.Queue
	store               store.Store

	stoppedC  chan struct{}
	stopC     chan struct{}
	httpstopC chan struct{}
	httpdoneC chan struct{}

	// used to coordinate shudown
	// lock should be used only in stop(), all other functions should use Rlock
	stopMu sync.RWMutex

	// waitProp waits for all proposals to be teminated before shutting down the node
	waitProp sync.WaitGroup
}

var (
	defaultSnapshotCount   uint64 = 10000
	snapshotCatchUpEntrisN uint64 = 10000
)

// TODO (upccup) result should contain more info
//type applyResult struct {
//	resp proto.Message
//	err  error
//}

type applyResult struct {
	resp string
	err  error
}

type InternalRaftRequest struct {
	ID    uint64
	Value string
}

func NewNode(id int, peers []string, store store.Store) *Node {
	n := Node{
		id:          id,
		peers:       peers,
		waldir:      fmt.Sprintf("node-%d", id),
		snapdir:     fmt.Sprintf("node-%d-snap", id),
		raftStorage: raft.NewMemoryStorage(),
		snapCount:   defaultSnapshotCount,
		stopC:       make(chan struct{}),
		httpstopC:   make(chan struct{}),
		httpdoneC:   make(chan struct{}),
		stoppedC:    make(chan struct{}),
		store:       store,
	}

	n.leadershipBroadcast = watch.NewQueue()
	n.ticker = clock.NewClock().NewTicker(time.Second)
	n.reqIDGen = idutil.NewGenerator(uint16(n.id), time.Now())
	n.wait = newWait()

	return &n
}

func (n *Node) StartRaft(ctx context.Context) error {
	if !fileutil.Exist(n.snapdir) {
		if err := os.Mkdir(n.snapdir, 0755); err != nil {
			return err
		}
	}

	n.snapshotter = snap.New(n.snapdir)

	oldwal := wal.Exist(n.waldir)
	wal, err := n.replayWAL()
	if err != nil {
		return err
	}

	n.wal = wal

	startPeers := make([]raft.Peer, len(n.peers))
	for i := range startPeers {
		startPeers[i] = raft.Peer{ID: uint64(i + 1)}
	}

	n.Config = &raft.Config{
		ID:              uint64(n.id),
		ElectionTick:    10,
		HeartbeatTick:   2,
		Storage:         n.raftStorage,
		MaxSizePerMsg:   1024 * 1024,
		MaxInflightMsgs: 256,
	}

	if oldwal {
		n.raftNode = raft.RestartNode(n.Config)
	} else {
		n.raftNode = raft.StartNode(n.Config, startPeers)
	}

	ss := &stats.ServerStats{}
	ss.Initialize()

	n.transport = &rafthttp.Transport{
		ID:          types.ID(n.id),
		ClusterID:   0x1000,
		Raft:        n,
		ServerStats: ss,
		LeaderStats: stats.NewLeaderStats(strconv.Itoa(n.id)),
		ErrorC:      make(chan error),
	}

	n.transport.Start()

	for i := range n.peers {
		if i+1 != n.id {
			n.transport.AddPeer(types.ID(i+1), []string{n.peers[i]})
		}
	}

	if err := n.serveRaft(); err != nil {
		return nil
	}

	snap, err := n.raftStorage.Snapshot()
	if err != nil {
		return err
	}

	n.confState = snap.Metadata.ConfState
	n.snapshotIndex = snap.Metadata.Index
	n.appliedIndex = snap.Metadata.Index

	atomic.StoreUint32(&n.isMember, 1)

	go func() {
		if err := n.Run(ctx); err != nil {
			log.L.Println("raft: run received loop got error: ", err)
		}
	}()

	return nil
}

func (n *Node) serveRaft() error {
	url, err := url.Parse(n.peers[n.id-1])
	if err != nil {
		return err
	}

	ln, err := newStoppableListener(url.Host, n.httpstopC)
	if err != nil {
		return err
	}

	go func() {
		err := (&http.Server{Handler: n.transport.Handler()}).Serve(ln)
		select {
		case <-n.httpstopC:
		default:
			log.L.Fatalf("serveRaft: failed to serve rafthttp %v", err)
		}

		close(n.httpstopC)
	}()

	return nil
}

// Run is the main loop for a raft node it goes along the state machine
// action on the messages received from other raft nodes in the cluster
// TODO(upccup)
// Before running the main loop it first starts the raft node based on saved
// cluster state. If no saved stater exists. It start a single-node cluster
func (n *Node) Run(ctx context.Context) error {
	ctx = log.WithLogger(ctx, logrus.WithField("raft_id", fmt.Sprintf("%x", n.id)))

	ctx, cancel := context.WithCancel(ctx)

	// nodeRemoved indicates that node was stopped due its removal
	nodeRemoved := false
	defer func() {
		cancel()
		n.stop(ctx)

		if nodeRemoved {
			// TODO(): remove wal and snapshot data
			log.G(ctx).Info("node have been removed")
		}
	}()

	wasLeader := false

	for {
		select {
		case <-n.ticker.C():
			n.raftNode.Tick()
		case rd := <-n.raftNode.Ready():
			n.wal.Save(rd.HardState, rd.Entries)

			if !raft.IsEmptySnap(rd.Snapshot) {
				n.saveSnap(rd.Snapshot)
				n.raftStorage.ApplySnapshot(rd.Snapshot)
				n.publishSnapshot(rd.Snapshot)
			}

			if rd.SoftState != nil {
				if wasLeader && rd.SoftState.RaftState != raft.StateLeader {
					wasLeader = false
					if atomic.LoadUint32(&n.signalledLeadership) == 1 {
						atomic.StoreUint32(&n.signalledLeadership, 0)
						n.leadershipBroadcast.Publish(IsFollower)
					}

					n.wait.cancelAll()
				} else if !wasLeader && rd.SoftState.RaftState == raft.StateLeader {
					wasLeader = true
				}
			}

			n.raftStorage.Append(rd.Entries)
			n.transport.Send(rd.Messages)

			if ok := n.publishEntries(n.entriesToApply(rd.CommittedEntries)); !ok {
				return errors.New("publishEntries failed")
			}

			n.maybeTriggerSnapshot()

			if wasLeader && atomic.LoadUint32(&n.signalledLeadership) != 1 {
				if n.caughtUp() {
					atomic.StoreUint32(&n.signalledLeadership, 1)
					n.leadershipBroadcast.Publish(IsLeader)
				}
			}

			n.raftNode.Advance()

		case err := <-n.transport.ErrorC:
			return err

		case <-n.stopC:
			n.stop(ctx)
			return nil
		}
	}
}

func (n *Node) LoadSanpshot() {
	snapshot, err := n.snapshotter.Load()
	if err == snap.ErrNoSnapshot {
		return
	}

	if err != nil && err != snap.ErrNoSnapshot {
		log.L.Panic(err)
	}

	log.L.Printf("loading snapshot at term %d and index %d", snapshot.Metadata.Term, snapshot.Metadata.Index)

	//TODO recover form snapshot
	//if err := s.revocerFromSnapshot(snapshot.Data); err != nil {
	//	log.L.Panic(err)
	//}
}

func (n *Node) caughtUp() bool {
	lastIndex, _ := n.raftStorage.LastIndex()
	return n.appliedIndex >= lastIndex
}

func (n *Node) WithContext(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-ctx.Done():
		case <-n.stoppedC:
			cancel()
		}
	}()

	return ctx, cancel
}

func (n *Node) canSubmitProposal() bool {
	select {
	case <-n.stoppedC:
		return false
	default:
		return true
	}
}

// TODO (upccup) raft message need common interface
//func (n *Node) ProposeValue(ctx context.Context, storeAction []ztypes.StoreAction, cb func()) error {
func (n *Node) ProposeValue(ctx context.Context, value string, cb func()) error {
	ctx, cancel := n.WithContext(ctx)
	defer cancel()

	_, err := n.processInternalRaftRequest(ctx, InternalRaftRequest{Value: value}, cb)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) processInternalRaftRequest(ctx context.Context, r InternalRaftRequest, cb func()) (*string, error) {
	n.stopMu.RLock()
	if !n.canSubmitProposal() {
		n.stopMu.RUnlock()
		return nil, ErrStopped
	}

	n.waitProp.Add(1)
	defer n.waitProp.Done()
	n.stopMu.RUnlock()

	r.ID = n.reqIDGen.Next()

	// this must be derived from the context which is cancelled by stop()
	// to avoid a deadlock on shutdown
	waitCtx, cancel := n.WithContext(ctx)

	ch := n.wait.register(r.ID, cb, cancel)

	// do this check after calling register to avoid a race
	if atomic.LoadUint32(&n.signalledLeadership) != 1 {
		n.wait.cancel(r.ID)
		return nil, ErrLostLeadership
	}

	data, err := json.Marshal(r)
	if err != nil {
		n.wait.cancel(r.ID)
		return nil, err
	}

	if len(data) > MaxTransactionBytes {
		n.wait.cancel(r.ID)
		return nil, ErrRequestToolarge
	}

	err = n.raftNode.Propose(waitCtx, data)
	if err != nil {
		n.wait.cancel(r.ID)
		return nil, err
	}

	select {
	case x := <-ch:
		res := x.(*applyResult)
		return &res.resp, res.err
	case <-waitCtx.Done():
		return nil, ErrLostLeadership
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (n *Node) SubscribeLeadership() (q chan events.Event, cancel func()) {
	return n.leadershipBroadcast.Watch()
}

func (n *Node) saveSnap(snap raftpb.Snapshot) error {
	if err := n.snapshotter.SaveSnap(snap); err != nil {
		return err
	}

	walSnap := walpb.Snapshot{
		Index: snap.Metadata.Index,
		Term:  snap.Metadata.Term,
	}

	if err := n.wal.SaveSnapshot(walSnap); err != nil {
		return err
	}

	return n.wal.ReleaseLockTo(snap.Metadata.Index)
}

func (n *Node) entriesToApply(ents []raftpb.Entry) []raftpb.Entry {
	var nents []raftpb.Entry

	if len(ents) == 0 {
		return nents
	}

	firstIdx := ents[0].Index

	if firstIdx > n.appliedIndex+1 {
		log.L.Fatalf("first index of committed entry [%d] should <= progress appliedIndex [%d] 1",
			firstIdx, n.appliedIndex)
	}

	if n.appliedIndex-firstIdx+1 <= uint64(len(ents)) {
		nents = ents[n.appliedIndex-firstIdx+1:]
	}

	return nents
}

func (n *Node) publishEntries(ents []raftpb.Entry) bool {
	for i := range ents {
		switch ents[i].Type {
		case raftpb.EntryNormal:
			if len(ents[i].Data) == 0 {
				break
			}

			var r InternalRaftRequest
			if err := json.Unmarshal(ents[i].Data, &r); err != nil {
				log.L.Errorf("store data got error", err.Error())
				return false
			}

			//TODO (upccup) store data to database
			//n.store.PutKeyValue(id, []byte(id))

			if !n.wait.trigger(r.ID, &applyResult{resp: r.Value, err: nil}) {
				n.wait.cancelAll()
			}

		case raftpb.EntryConfChange:
			var cc raftpb.ConfChange
			cc.Unmarshal(ents[i].Data)

			n.raftNode.ApplyConfChange(cc)

			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				if len(cc.Context) > 0 {
					n.transport.AddPeer(types.ID(cc.NodeID), []string{string(cc.Context)})
				}

			case raftpb.ConfChangeRemoveNode:
				if cc.NodeID == uint64(n.id) {
					log.L.Println("I've been removed from the cluster! Shutting down")
				}

				n.transport.RemovePeer(types.ID(cc.NodeID))
			}
		}

		// after commit update appliedIndex
		n.appliedIndex = ents[i].Index

		// sepcial nil commit to signal replay has finished
		if ents[i].Index == n.lastIndex {
			n.LoadSanpshot()
		}
	}

	return true
}

// returns a WAL ready for reading
func (n *Node) openWAL() (*wal.WAL, error) {
	if !wal.Exist(n.waldir) {
		if err := os.Mkdir(n.waldir, 0755); err != nil {
			return nil, err
		}

		w, err := wal.Create(n.waldir, nil)
		if err != nil {
			return nil, err
		}

		w.Close()
	}

	w, err := wal.Open(n.waldir, walpb.Snapshot{})
	if err != nil {
		return nil, err
	}

	return w, nil
}

// replays WAL entries into the raft intance
func (n *Node) replayWAL() (*wal.WAL, error) {
	w, err := n.openWAL()
	if err != nil {
		return nil, err
	}

	_, st, ents, err := w.ReadAll()
	if err != nil {
		return nil, err
	}

	// appent to storage so raft starts at the right place log
	n.raftStorage.Append(ents)

	if len(ents) > 0 {
		n.lastIndex = ents[len(ents)-1].Index
	} else {
		n.LoadSanpshot()
	}

	n.raftStorage.SetHardState(st)

	return w, nil
}

// checks if the raft node has effectively joined a cluster of exiting member
func (n *Node) IsMember() bool {
	return atomic.LoadUint32(&n.isMember) == 1
}

// checks if we are the leader or nor, without the protection of lock
func (n *Node) isLeader() bool {
	if !n.IsMember() {
		return false
	}

	if n.Status().Lead == n.Config.ID {
		return true
	}

	return false
}

// checks if we are the leader of not, with the protection of lock
func (n *Node) IsLeader() bool {
	n.stopMu.RLock()
	defer n.stopMu.RUnlock()
	return n.isLeader()
}

// returns status of underlying etcd.Node
func (n *Node) Status() raft.Status {
	return n.raftNode.Status()
}

// returns the id of the leader, without the protection of lock and membership check, so it's caller task
func (n *Node) leader() uint64 {
	return n.Status().Lead
}

// returns the id of leader, with the protection of lock
func (n *Node) Leader() (uint64, error) {
	n.stopMu.RLock()
	defer n.stopMu.RUnlock()

	if !n.IsMember() {
		return raft.None, ErrNoRaftMember
	}

	leader := n.leader()
	if leader == raft.None {
		return raft.None, ErrNoClusterLeader
	}

	return leader, nil
}

// close http and close all channels and stop raft
func (n *Node) stop(ctx context.Context) {
	n.stopHTTP()
	n.leadershipBroadcast.Close()
	n.ticker.Stop()
	n.raftNode.Stop()
	atomic.StoreUint32(&n.isMember, 0)
}

func (n *Node) stopHTTP() {
	n.transport.Stop()
	close(n.httpstopC)
	<-n.httpdoneC
}

func (n *Node) publishSnapshot(snapshotToSave raftpb.Snapshot) {
	if raft.IsEmptySnap(snapshotToSave) {
		return
	}

	log.L.Printf("publishing snapshot at index %d", n.snapshotIndex)
	defer log.L.Printf("finished publish snapshot at index", n.snapshotIndex)

	if snapshotToSave.Metadata.Index >= n.appliedIndex {
		log.L.Errorf("publish snapshot: snapshot index [%d] should > progress.appliedIndex [%d] + 1",
			snapshotToSave.Metadata.Index, n.appliedIndex)
		return
	}

	n.LoadSanpshot()

	n.confState = snapshotToSave.Metadata.ConfState
	n.snapshotIndex = snapshotToSave.Metadata.Index
	n.appliedIndex = snapshotToSave.Metadata.Index
}

func (n *Node) maybeTriggerSnapshot() {
	if n.appliedIndex-n.snapshotIndex <= n.snapCount {
		return
	}

	log.L.Printf("maybeTriggerSnapshot: start snapshot [applied index %d | last snapshot index %d]",
		n.appliedIndex, n.snapshotIndex)

	//TODO add getSanpshot() in store
	//data, err := n.store.GetSnapshot()
	//if err != nil {
	//	log.L.Error("maybeTriggerSnapshot: get snapshot failed error: ", err)
	//	return
	//}

	data := []byte{}
	snap, err := n.raftStorage.CreateSnapshot(n.appliedIndex, &n.confState, data)
	if err != nil {
		log.L.Error("maybeTriggerSnapshot: create snapshot failed error: ", err)
		return
	}

	if err := n.saveSnap(snap); err != nil {
		log.L.Error("maybeTriggerSnapshot: save snapshot failed error: ", err)
		return
	}

	compactIndex := uint64(1)

	if n.appliedIndex > snapshotCatchUpEntrisN {
		compactIndex = n.appliedIndex - snapshotCatchUpEntrisN
	}

	if err := n.raftStorage.Compact(compactIndex); err != nil {
		log.L.Error("maybeTriggerSnapshot: compact snapshot failed error: ", err)
		return
	}

	log.L.Printf("maybeTriggerSnapshot: Compact log at inex %d", compactIndex)
	n.snapshotIndex = n.appliedIndex
}

func (n *Node) Process(ctx context.Context, m raftpb.Message) error {
	return n.raftNode.Step(ctx, m)
}

func (n *Node) IsIDRemoved(id uint64) bool                           { return false }
func (n *Node) ReportUnreachable(id uint64)                          {}
func (n *Node) ReportSnapshot(id uint64, status raft.SnapshotStatus) {}
