package upstream

import (
	"encoding/json"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Sessions
type Sessions struct {
	m            map[string]*session // ip -> session
	sync.RWMutex                     // protect m
	stopCh       chan struct{}       // quit
	gcInterval   time.Duration       // gc interval
	timeout      time.Duration       // session timeout
}

type session struct {
	*Target
	UpdatedAt time.Time `json:"updated_at"`
}

func newSessions() *Sessions {
	b := &Sessions{
		m:          make(map[string]*session),
		stopCh:     make(chan struct{}),
		gcInterval: time.Second * 10,
		timeout:    time.Hour * 1,
	}

	go b.gc()
	return b
}

func (s *Sessions) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.m)
}

func (s *Sessions) get(ip string) *Target {
	s.RLock()
	defer s.RUnlock()
	sess, ok := s.m[ip]
	if !ok {
		return nil
	}
	return sess.Target
}

func (s *Sessions) update(ip string, t *Target) {
	s.Lock()
	s.m[ip] = &session{t, time.Now()}
	s.Unlock()
}

func (s *Sessions) remove(taskID string) {
	s.Lock()
	for k, v := range s.m {
		if v.TaskID == taskID {
			delete(s.m, k)
		}
	}
	s.Unlock()
}

func (s *Sessions) gc() {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()

	for {
		select {

		case <-ticker.C:
			s.Lock()
			for key, session := range s.m {
				if session.UpdatedAt.Before(time.Now().Add(-s.timeout)) {
					log.Printf("clean up outdated session: %s -> %s", key, session.TaskID)
					delete(s.m, key)
				}
			}
			s.Unlock()

		case <-s.stopCh:
			return
		}
	}
}

// stop gc and clean up
func (s *Sessions) stop() {
	close(s.stopCh)
	s.m = map[string]*session{} // gc friendly
}
