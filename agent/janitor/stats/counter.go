package stats

import (
	"encoding/json"
	"time"
)

var (
	stats         *Stats
	rateFreshIntv = time.Second * 2  // rate calculation interval
	gcIntv        = time.Second * 10 // interval to scan & clean up removal-marked backend counter
)

func init() {
	stats = &Stats{
		Global:       &GlobalCounter{startedAt: time.Now()},
		Upstream:     make(UpstreamCounter),
		inGlbCh:      make(chan *DeltaGlb, 1024),
		inBackendCh:  make(chan *DeltaBackend, 1024),
		delBackendCh: make(chan *DeltaBackend, 128),
		queryCh:      make(chan chan Stats),
	}

	go stats.runCounters()
}

// Stats holds all of statistics data.
type Stats struct {
	Global   *GlobalCounter  `json:"global"`   // global counter
	Upstream UpstreamCounter `json:"upstream"` // upstream -> backend -> counter

	inGlbCh      chan *DeltaGlb     // new global counter delta received
	inBackendCh  chan *DeltaBackend // new upstream/backend counter delta received
	delBackendCh chan *DeltaBackend // removal signal upstream->backend counter delta
	queryCh      chan chan Stats
}

// GlobalCounter hold current global statistics
type GlobalCounter struct {
	RxBytes  uint64 `json:"rx_bytes"`      // nb of received bytes
	TxBytes  uint64 `json:"tx_bytes"`      // nb of transmitted bytes
	Requests uint64 `json:"requests"`      // nb of client requests
	Fails    uint64 `json:"fails"`         // nb of failed requests
	RxRate   uint   `json:"rx_rate"`       // received bytes / second
	TxRate   uint   `json:"tx_rate"`       // transmitted bytes / second
	ReqRate  uint   `json:"requests_rate"` // requests / second
	FailRate uint   `json:"fails_rate"`    // failed requests / second

	lastRx   uint64 // used for calculate rate per second
	lastTx   uint64
	lastReq  uint64
	lastFail uint64
	freshed  bool

	startedAt time.Time
}

type GlobalCounterAlias GlobalCounter

func (c *GlobalCounter) MarshalJSON() ([]byte, error) {
	var wrapper struct {
		GlobalCounterAlias
		Uptime string `json:"uptime"`
	}

	wrapper.GlobalCounterAlias = GlobalCounterAlias(*c)
	wrapper.Uptime = time.Now().Sub(c.startedAt).String()
	return json.Marshal(wrapper)
}

// UpstreamCounter hold upstream & backends current statistics
type UpstreamCounter map[string]map[string]*BackendCounter

// BackendCounter hold one upstream-backend's current statistics
type BackendCounter struct {
	ActiveClients uint   `json:"active_clients"` // active clients
	RxBytes       uint64 `json:"rx_bytes"`       // nb of received bytes
	TxBytes       uint64 `json:"tx_bytes"`       // nb of transmitted bytes
	Requests      uint64 `json:"requests"`       // nb of requests
	RxRate        uint   `json:"rx_rate"`        // received bytes / second
	TxRate        uint   `json:"tx_rate"`        // transmitted bytes / second
	ReqRate       uint   `json:"requests_rate"`  // requests / second

	lastRx  uint64 // used for calculate rate per second
	lastTx  uint64
	lastReq uint64
	freshed bool

	startedAt time.Time

	removed bool // removal flag, actually removed by gc() until `ActiveClients` decreased to 0
}

type BackendCounterAlias BackendCounter

func (c *BackendCounter) MarshalJSON() ([]byte, error) {
	var wrapper struct {
		BackendCounterAlias
		Uptime string `json:"uptime"`
	}

	wrapper.BackendCounterAlias = BackendCounterAlias(*c)
	wrapper.Uptime = time.Now().Sub(c.startedAt).String()
	return json.Marshal(wrapper)
}

type DeltaBackend struct {
	Uid string
	Bid string
	Ac  int
	Rx  uint64
	Tx  uint64
	Req uint64
}

type DeltaGlb struct {
	Rx   uint64
	Tx   uint64
	Req  uint64
	Fail uint64
}

func Get() *Stats {
	ch := make(chan Stats)
	stats.queryCh <- ch
	s := <-ch
	return &s
}

func UpstreamStats() UpstreamCounter {
	return Get().Upstream

}

func Incr(dbackend *DeltaBackend, dglb *DeltaGlb) {
	if dbackend != nil {
		stats.inBackendCh <- dbackend
	}
	if dglb != nil {
		stats.inGlbCh <- dglb
	}
}

func Del(ups, backend string) {
	stats.delBackendCh <- &DeltaBackend{Uid: ups, Bid: backend}
}

func (c *Stats) runCounters() {
	freshTicker := time.NewTicker(rateFreshIntv)
	defer freshTicker.Stop()

	gcTicker := time.NewTicker(gcIntv)
	defer gcTicker.Stop()

	for {
		select {
		case <-freshTicker.C:
			c.freshRate()
		case <-gcTicker.C:
			c.gc()
		case d := <-c.inBackendCh:
			c.updateBackend(d)
		case d := <-c.inGlbCh:
			c.updateGlb(d)
		case d := <-c.delBackendCh:
			c.removeBackend(d)
		case ch := <-c.queryCh:
			ch <- *c
		}
	}
}

func (c *Stats) freshRate() {
	c.Global.freshRate()

	for _, ups := range c.Upstream {
		for _, backend := range ups {
			backend.freshRate()
		}
	}
}

// fresh global counter
func (c *GlobalCounter) freshRate() {
	if !c.freshed {
		c.RxRate = 0
		c.TxRate = 0
		c.ReqRate = 0
		c.FailRate = 0
		return
	}

	var (
		nRx   = c.RxBytes - c.lastRx
		nTx   = c.TxBytes - c.lastTx
		nReq  = c.Requests - c.lastReq
		nFail = c.Fails - c.lastFail
		intv  = uint64(rateFreshIntv.Seconds())
	)

	c.RxRate = uint(nRx / intv)
	c.TxRate = uint(nTx / intv)
	c.ReqRate = uint(nReq / intv)
	c.FailRate = uint(nFail / intv)

	c.lastRx = c.RxBytes
	c.lastTx = c.TxBytes
	c.lastReq = c.Requests
	c.lastFail = c.Fails

	c.freshed = false // mark as consumed
}

// fresh backend counter
func (c *BackendCounter) freshRate() {
	if !c.freshed {
		c.RxRate = 0
		c.TxRate = 0
		c.ReqRate = 0
		return
	}

	var (
		nRx  = c.RxBytes - c.lastRx
		nTx  = c.TxBytes - c.lastTx
		nReq = c.Requests - c.lastReq
		intv = uint64(rateFreshIntv.Seconds())
	)

	c.RxRate = uint(nRx / intv)
	c.TxRate = uint(nTx / intv)
	c.ReqRate = uint(nReq / intv)

	c.lastRx = c.RxBytes
	c.lastTx = c.TxBytes
	c.lastReq = c.Requests

	c.freshed = false // mark as consumed
}

// gc actually delete removal-marked backend counters
// until the `active-clients` decreased to zero
func (c *Stats) gc() {
	for uid, ups := range c.Upstream {

		for bid, backend := range ups {
			if !backend.removed {
				continue
			}
			if backend.ActiveClients <= 0 {
				delete(ups, bid)
			}
		}

		if len(ups) == 0 {
			delete(c.Upstream, uid)
		}
	}
}

func (c *Stats) updateGlb(d *DeltaGlb) {
	c.Global.RxBytes += d.Rx
	c.Global.TxBytes += d.Tx
	c.Global.Requests += d.Req
	c.Global.Fails += d.Fail
	c.Global.freshed = true
}

func (c *Stats) updateBackend(d *DeltaBackend) {
	var (
		uid = d.Uid
		bid = d.Bid
	)

	if uid == "" || bid == "" {
		return
	}

	if _, ok := c.Upstream[uid]; !ok {
		c.Upstream[uid] = make(map[string]*BackendCounter)
	}
	ups := c.Upstream[uid]

	if _, ok := ups[bid]; !ok {
		ups[bid] = &BackendCounter{
			startedAt: time.Now(),
		}
	}
	backend := ups[bid]

	backend.ActiveClients += uint(d.Ac)
	if backend.ActiveClients < 0 {
		backend.ActiveClients = 0
	}

	if n := d.Rx; n > 0 {
		backend.RxBytes += n
	}
	if n := d.Tx; n > 0 {
		backend.TxBytes += n
	}
	if n := d.Req; n > 0 {
		backend.Requests += n
	}

	backend.freshed = true
}

// note: removeBackend() only mark the removal flag on specified backend counter,
// counter will be actually removed by gc() until its `active-clients` decreased to zero
func (c *Stats) removeBackend(d *DeltaBackend) {
	var (
		uid = d.Uid
		bid = d.Bid
	)

	if uid == "" || bid == "" {
		return
	}
	if _, ok := c.Upstream[uid]; !ok {
		return
	}
	ups := c.Upstream[uid]

	if backend, ok := ups[bid]; ok {
		backend.removed = true
	}
}
