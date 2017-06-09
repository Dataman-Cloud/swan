package janitor

import (
	"encoding/json"
	"time"
)

var (
	rateFreshIntv = time.Second * 2  // rate calculation interval
	gcIntv        = time.Second * 10 // interval to scan & clean up removal-marked task counter
)

// Stats holds all of statistics data.
type Stats struct {
	Global *GlobalCounter `json:"global"` // global counter
	App    AppCounter     `json:"app"`    // app -> task -> counter

	inGlbCh  chan *deltaGlb // new global counter delta received
	inAppCh  chan *deltaApp // new app counter delta received
	delAppCh chan *deltaApp // removal signal app->task counter delta
}

// GlobalCounter hold current global statistics
type GlobalCounter struct {
	RxBytes  uint64 `json:"rx_bytes"`      // nb of received bytes
	TxBytes  uint64 `json:"tx_bytes"`      // nb of transmitted bytes
	Requests uint64 `json:"requests"`      // nb of client requests
	Fails    uint64 `json:"fails"`         // nb of failed requesets
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

// AppCounter hold app current statistics
type AppCounter map[string]map[string]*TaskCounter

// TaskCounter hold one app-task's current statistics
type TaskCounter struct {
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

type TaskCounterAlias TaskCounter

func (c *TaskCounter) MarshalJSON() ([]byte, error) {
	var wrapper struct {
		TaskCounterAlias
		Uptime string `json:"uptime"`
	}

	wrapper.TaskCounterAlias = TaskCounterAlias(*c)
	wrapper.Uptime = time.Now().Sub(c.startedAt).String()
	return json.Marshal(wrapper)
}

type deltaApp struct {
	aid string
	tid string
	ac  int
	rx  uint64
	tx  uint64
	req uint64
}

type deltaGlb struct {
	rx   uint64
	tx   uint64
	req  uint64
	fail uint64
}

func newStats() *Stats {
	c := &Stats{
		Global: &GlobalCounter{
			startedAt: time.Now(),
		},
		App:      make(AppCounter),
		inGlbCh:  make(chan *deltaGlb, 1024),
		inAppCh:  make(chan *deltaApp, 1024),
		delAppCh: make(chan *deltaApp, 128),
	}

	go c.runCounters()
	return c
}

func (c *Stats) incr(dapp *deltaApp, dglb *deltaGlb) {
	if dapp != nil {
		c.inAppCh <- dapp
	}
	if dglb != nil {
		c.inGlbCh <- dglb
	}
}

func (c *Stats) del(aid, tid string) {
	c.delAppCh <- &deltaApp{aid: aid, tid: tid}
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
		case d := <-c.inAppCh:
			c.updateApp(d)
		case d := <-c.inGlbCh:
			c.updateGlb(d)
		case d := <-c.delAppCh:
			c.removeApp(d)
		}
	}
}

func (c *Stats) freshRate() {
	c.Global.freshRate()

	for _, app := range c.App {
		for _, task := range app {
			task.freshRate()
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

// fresh task counter
func (c *TaskCounter) freshRate() {
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

// gc actually delete removal-marked task counters
// until the `active-clients` decreased to zero
func (c *Stats) gc() {
	for aid, app := range c.App {

		for tid, task := range app {
			if !task.removed {
				continue
			}
			if task.ActiveClients <= 0 {
				delete(app, tid)
			}
		}

		if len(app) == 0 {
			delete(c.App, aid)
		}
	}
}

func (c *Stats) updateGlb(d *deltaGlb) {
	c.Global.RxBytes += d.rx
	c.Global.TxBytes += d.tx
	c.Global.Requests += d.req
	c.Global.Fails += d.fail
	c.Global.freshed = true
}

func (c *Stats) updateApp(d *deltaApp) {
	if d.aid == "" || d.tid == "" {
		return
	}

	if _, ok := c.App[d.aid]; !ok {
		c.App[d.aid] = make(map[string]*TaskCounter)
	}
	app := c.App[d.aid]

	if _, ok := app[d.tid]; !ok {
		app[d.tid] = &TaskCounter{
			startedAt: time.Now(),
		}
	}
	task := app[d.tid]

	task.ActiveClients += uint(d.ac)
	if task.ActiveClients < 0 {
		task.ActiveClients = 0
	}

	if n := d.rx; n > 0 {
		task.RxBytes += n
	}
	if n := d.tx; n > 0 {
		task.TxBytes += n
	}
	if n := d.req; n > 0 {
		task.Requests += n
	}

	task.freshed = true
}

// note: removeApp() only mark the removal flag on specified task counter,
// counter will be actually removed by gc() until its `active-clients` decreased to zero
func (c *Stats) removeApp(d *deltaApp) {
	if d.aid == "" || d.tid == "" {
		return
	}
	if _, ok := c.App[d.aid]; !ok {
		return
	}
	app := c.App[d.aid]

	if task, ok := app[d.tid]; ok {
		task.removed = true
	}
}
