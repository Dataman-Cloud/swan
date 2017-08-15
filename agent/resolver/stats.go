package resolver

import (
	"encoding/json"
	"sync"
	"time"
)

type Stats struct {
	sync.RWMutex                     // protect the followings two
	Global       *Counter            `json:"global"`
	Parents      map[string]*Counter `json:"parents"` // parent -> counter

	startedAt time.Time
}

type StatsAlias Stats

type Counter struct {
	Requests  uint64 `json:"requests"`  // nb of client requests
	Fails     uint64 `json:"fails"`     // nb of failed requests
	Authority uint64 `json:"authority"` // nb of authority requests
	Forward   uint64 `json:"forward"`   // nb of forward requests
	TypeA     uint64 `json:"type_a"`    // nb of A requests
	TypeSRV   uint64 `json:"type_srv"`  // nb of SRV requests
}

func newStats() *Stats {
	return &Stats{
		Global:    new(Counter),
		Parents:   make(map[string]*Counter),
		startedAt: time.Now(),
	}
}

func (s *Stats) MarshalJSON() ([]byte, error) {
	var wrapper struct {
		StatsAlias        // prevent marshal OOM
		Uptime     string `json:"uptime"`
	}

	wrapper.StatsAlias = StatsAlias(*s)
	wrapper.Uptime = time.Now().Sub(s.startedAt).String()
	return json.Marshal(wrapper)
}

func (s *Stats) Get() *Stats {
	s.RLock()
	defer s.RUnlock()
	return s
}

func (s *Stats) Incr(p string, delta *Counter) {
	s.Lock()
	defer s.Unlock()

	s.Global.Requests += delta.Requests
	s.Global.Fails += delta.Fails
	s.Global.Authority += delta.Authority
	s.Global.Forward += delta.Forward
	s.Global.TypeA += delta.TypeA
	s.Global.TypeSRV += delta.TypeSRV

	// skip non authority delta counter
	if delta.Authority == 0 {
		return
	}

	if _, ok := s.Parents[p]; !ok {
		s.Parents[p] = new(Counter)
	}
	c := s.Parents[p]

	c.Requests += delta.Requests
	c.Fails += delta.Fails
	c.Authority += delta.Authority
	c.Forward += delta.Forward
	c.TypeA += delta.TypeA
	c.TypeSRV += delta.TypeSRV
}

func (s *Stats) Del(p string) {
	s.Lock()
	defer s.Unlock()

	delete(s.Parents, p)
}
