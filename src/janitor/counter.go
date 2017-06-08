package janitor

import (
	"encoding/json"
	"time"
)

// Stats holds all of statistics data.
type Stats struct {
	Uptime   string                             `json:"uptime"`
	Global   *GlobalCounter                     `json:"global"`
	App      map[string]map[string]*TaskCounter `json:"app"` // app -> task -> counter
	inGlbCh  chan *deltaGlb                     // new global counter delta received
	inAppCh  chan *deltaApp                     // new app counter delta received
	delAppCh chan *deltaApp                     // removal signal app->task counter delta
	startAt  time.Time
}

type StatsAlias Stats

// GlobalCounter hold current global statistics
type GlobalCounter struct {
	RxBytes  uint64 `json:"rx_bytes"` // nb of received bytes
	TxBytes  uint64 `json:"tx_bytes"` // nb of transmitted bytes
	Requests uint64 `json:"requests"` // nb of client requests
	Fails    uint64 `json:"fails"`    // nb of failed requesets
}

// TaskCounter hold one app-task's current statistics
type TaskCounter struct {
	ActiveClients uint   `json:"active_clients"` // active clients
	RxBytes       uint64 `json:"rx_bytes"`       // nb of received bytes
	TxBytes       uint64 `json:"tx_bytes"`       // nb of transmitted bytes
	Requests      uint64 `json:"requests"`       // nb of requests
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
		Global:   &GlobalCounter{},
		App:      make(map[string]map[string]*TaskCounter),
		inGlbCh:  make(chan *deltaGlb, 1024),
		inAppCh:  make(chan *deltaApp, 1024),
		delAppCh: make(chan *deltaApp, 128),
		startAt:  time.Now(),
	}

	go c.runCounters()
	return c
}

func (c *Stats) MarshalJSON() ([]byte, error) {
	a := StatsAlias(*c)
	a.Uptime = time.Now().Sub(a.startAt).String()
	return json.Marshal(a)
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
	for {
		select {
		case d := <-c.inAppCh:
			c.updateApp(d)
		case d := <-c.inGlbCh:
			c.updateGlb(d)
		case d := <-c.delAppCh:
			c.removeApp(d)
		}
	}
}

func (c *Stats) updateGlb(d *deltaGlb) {
	c.Global.RxBytes += d.rx
	c.Global.TxBytes += d.tx
	c.Global.Requests += d.req
	c.Global.Fails += d.fail
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
		app[d.tid] = new(TaskCounter)
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
}

func (c *Stats) removeApp(d *deltaApp) {
	if d.aid == "" || d.tid == "" {
		return
	}
	if _, ok := c.App[d.aid]; !ok {
		return
	}
	app := c.App[d.aid]

	if _, ok := app[d.tid]; ok {
		delete(app, d.tid)
	}

	if len(app) == 0 {
		delete(c.App, d.aid)
	}
}
