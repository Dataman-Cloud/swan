package upstream

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var mgr *UpsManager

func init() {
	mgr = &UpsManager{
		Upstreams: make([]*Upstream, 0, 0),
	}
}

// Type & Method Definations ...
//
type UpsManager struct {
	Upstreams []*Upstream `json:"upstreams"`
	sync.RWMutex
}

type Upstream struct {
	Name     string     `json:"name"`     // uniq name
	Alias    string     `json:"alias"`    // advertised url
	Listen   string     `json:"listen"`   // listen addr
	Backends []*Backend `json:"backends"` // backend servers

	sessions *Sessions // runtime
	balancer Balancer  // runtime
}

func newUpstream(first *BackendCombined) *Upstream {
	return &Upstream{
		Name:     first.Upstream.Name,
		Alias:    first.Upstream.Alias,
		Listen:   first.Upstream.Listen,
		Backends: []*Backend{first.Backend},
		sessions: newSessions(),     // sessions store
		balancer: &weightBalancer{}, // default balancer
	}
}

func (u *Upstream) valid() error {
	if u == nil {
		return errors.New("nil upstream")
	}
	if u.Name == "" {
		return errors.New("upstream name required")
	}
	return nil
}

func (u *Upstream) search(name string) (int, *Backend) {
	for i, v := range u.Backends {
		if v.ID == name {
			return i, v
		}
	}
	return -1, nil
}

func (u *Upstream) tcpListen() string {
	if u.Listen == "" {
		return ""
	}

	ss := strings.Split(u.Listen, ":")
	if port := ss[len(ss)-1]; port != "" {
		return ":" + port
	}

	return ""
}

// Backend
type Backend struct {
	ID      string  `json:"id"`     // backend server id(name)
	IP      string  `json:"ip"`     // backend server ip
	Port    uint32  `json:"port"`   // backend server port
	Scheme  string  `json:"scheme"` // http / https, auto detect & setup by httpProxy
	Version string  `json:"version"`
	Weight  float64 `json:"weihgt"`
}

func (b *Backend) valid() error {
	if b == nil {
		return errors.New("nil backend")
	}
	if b.ID == "" {
		return errors.New("backend id required")
	}
	if b.IP == "" {
		return errors.New("backend ip required")
	}
	if b.Port == 0 {
		return errors.New("backend port required")
	}
	return nil
}

func (b *Backend) Addr() string {
	return fmt.Sprintf("%s:%d", b.IP, b.Port)
}

// BackendEvent
type BackendEvent struct {
	Action string // add/del/update
	*BackendCombined
}

func (ev *BackendEvent) String() string {
	return fmt.Sprintf("{%s upstream:%s backend:%s addr:%s:%d weight:%f}",
		ev.Action, ev.Upstream.Name, ev.Backend.ID,
		ev.Backend.IP, ev.Backend.Port, ev.Backend.Weight)
}

func (ev *BackendEvent) Format() {
	ev.Upstream.Listen = ev.Upstream.tcpListen() // rewrite Upstream.Listen
}

func BuildBackendEvent(act, ups, alias, listen, backend, ip, ver string, port uint32, weight float64) *BackendEvent {
	return &BackendEvent{
		Action: act,
		BackendCombined: &BackendCombined{
			Upstream: &Upstream{Name: ups, Alias: alias, Listen: listen},
			Backend:  &Backend{backend, ip, port, "", ver, weight},
		},
	}
}

// BackendCombined
type BackendCombined struct {
	*Upstream `json:"upstream"`
	*Backend  `json:"backend"`
}

func (cmb *BackendCombined) Valid() error {
	if cmb == nil {
		return errors.New("nil backend combined")
	}
	if err := cmb.Upstream.valid(); err != nil {
		return err
	}
	if err := cmb.Backend.valid(); err != nil {
		return err
	}
	if !strings.HasSuffix(cmb.Backend.ID, "-"+cmb.Upstream.Name) {
		return errors.New("backend name must be suffixed by upstream name")
	}
	return nil
}

// Exported Functions ....
//
func AllUpstreams() []*Upstream {
	mgr.RLock()
	defer mgr.RUnlock()
	return mgr.Upstreams
}

func AllSessions() map[string]*Sessions {
	mgr.RLock()
	defer mgr.RUnlock()

	ret := make(map[string]*Sessions)
	for _, u := range mgr.Upstreams {
		ret[u.Name] = u.sessions
	}
	return ret
}

func GetUpstream(ups string) *Upstream {
	mgr.RLock()
	defer mgr.RUnlock()

	_, u := getUpstreamByName(ups)
	return u
}

func UpsertBackend(cmb *BackendCombined) (onFirst bool, err error) {
	mgr.Lock()
	defer mgr.Unlock()

	var (
		ups     = cmb.Upstream.Name
		alias   = cmb.Upstream.Alias
		listen  = cmb.Upstream.Listen
		backend = cmb.Backend.ID
	)

	_, u := getUpstreamByName(ups)
	// add new upstream
	if u == nil {
		onFirst = true

		if i, _ := getUpstreamByAlias(alias); i >= 0 {
			err = fmt.Errorf("alias address [%s] conflict", alias)
			return
		}
		if i, _ := getUpstreamByListen(listen); i >= 0 {
			err = fmt.Errorf("listen address [%s] conflict", listen)
			return
		}

		mgr.Upstreams = append(mgr.Upstreams, newUpstream(cmb))
		return
	}

	_, b := u.search(backend)

	// add new backend
	if b == nil {
		u.Backends = append(u.Backends, cmb.Backend)
		return
	}

	// update upstream
	u.Alias = cmb.Upstream.Alias

	// update backend
	b.IP = cmb.Backend.IP
	b.Port = cmb.Backend.Port
	b.Scheme = cmb.Backend.Scheme
	b.Version = cmb.Backend.Version
	b.Weight = cmb.Backend.Weight

	return
}

func GetBackend(ups, backend string) *Backend {
	mgr.RLock()
	defer mgr.RUnlock()

	_, u := getUpstreamByName(ups)
	if u == nil {
		return nil
	}

	_, b := u.search(backend)
	return b
}

func RemoveBackend(cmb *BackendCombined) (onLast bool) {
	mgr.Lock()
	defer mgr.Unlock()

	var (
		ups     = cmb.Upstream.Name
		backend = cmb.Backend.ID
	)

	idxu, u := getUpstreamByName(ups)
	if u == nil {
		return
	}

	idxb, b := u.search(backend)
	if b == nil {
		return
	}

	// remove backend & session
	u.Backends = append(u.Backends[:idxb], u.Backends[idxb+1:]...)
	u.sessions.remove(backend)

	// remove empty upstream & stop sessions gc
	if len(u.Backends) == 0 {
		onLast = true
		u.sessions.stop()
		mgr.Upstreams = append(mgr.Upstreams[:idxu], mgr.Upstreams[idxu+1:]...)
	}

	return
}

// similar as lookup, but by upstream alias
func LookupAlias(remoteIP, alias string) *BackendCombined {
	mgr.RLock()
	_, u := getUpstreamByAlias(alias)
	mgr.RUnlock()

	if u == nil {
		return nil
	}

	return Lookup(remoteIP, u.Name, "")
}

// similar as lookup, but by upstream listen
func LookupListen(remoteIP, listen string) *BackendCombined {
	mgr.RLock()
	_, u := getUpstreamByListen(listen)
	mgr.RUnlock()

	if u == nil {
		return nil
	}

	return Lookup(remoteIP, u.Name, "")
}

// lookup select a suitable backend according by sessions & balancer
func Lookup(remoteIP, ups, backend string) *BackendCombined {
	var (
		u *Upstream
		b *Backend
	)

	if _, u = getUpstreamByName(ups); u == nil {
		return nil
	}

	defer func() {
		if b != nil {
			u.sessions.update(remoteIP, b)
		}
	}()

	// obtain specified backend
	if backend != "" {
		b = GetBackend(ups, backend)
		if b == nil {
			return nil
		}
		return &BackendCombined{u, b}
	}

	// obtain session by remoteIP
	if b = u.sessions.get(remoteIP); b != nil {
		return &BackendCombined{u, b}
	}

	// use balancer to obtain a new backend
	if b = nextBackend(ups); b != nil {
		return &BackendCombined{u, b}
	}

	return nil
}

func nextBackend(ups string) *Backend {
	mgr.RLock()
	defer mgr.RUnlock()

	_, u := getUpstreamByName(ups)
	if u == nil {
		return nil
	}

	return u.balancer.Next(u.Backends)
}

// note: must be called under protection of mutext lock
func getUpstreamByName(ups string) (int, *Upstream) {
	for i, v := range mgr.Upstreams {
		if v.Name == ups {
			return i, v
		}
	}
	return -1, nil
}

// note: must be called under protection of mutext lock
func getUpstreamByAlias(alias string) (int, *Upstream) {
	if alias == "" {
		return -1, nil
	}
	for i, v := range mgr.Upstreams {
		if v.Alias == alias {
			return i, v
		}
	}
	return -1, nil
}

// note: must be called under protection of mutext lock
func getUpstreamByListen(listen string) (int, *Upstream) {
	if listen == "" {
		return -1, nil
	}
	for i, v := range mgr.Upstreams {
		if v.Listen == listen {
			return i, v
		}
	}
	return -1, nil
}
