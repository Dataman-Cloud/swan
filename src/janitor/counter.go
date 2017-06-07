package janitor

import (
	"encoding/json"
	"time"
)

// Stats holds all of statistics data.
type Stats struct {
	Uptime  string                             `json:"uptime"`
	Global  *GlobalCounter                     `json:"global"`
	App     map[string]map[string]*TaskCounter `json:"app"` // app -> task -> counter
	inGlb   chan *deltaGlb                     // new global counter delta
	inApp   chan *deltaApp                     // new app counter delta
	startAt time.Time
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
}

type deltaApp struct {
	aid string
	tid string
	ac  int
	rx  uint64
	tx  uint64
}

type deltaGlb struct {
	tx   uint64
	rx   uint64
	req  uint64
	fail uint64
}

func newStats() *Stats {
	c := &Stats{
		Global:  &GlobalCounter{},
		App:     make(map[string]map[string]*TaskCounter),
		inGlb:   make(chan *deltaGlb, 1024),
		inApp:   make(chan *deltaApp, 1024),
		startAt: time.Now(),
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
		c.inApp <- dapp
	}
	if dglb != nil {
		c.inGlb <- dglb
	}
}

func (c *Stats) runCounters() {
	for {
		select {
		case d := <-c.inApp:
			c.updateApp(d)
		case d := <-c.inGlb:
			c.updateGlb(d)
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
}
