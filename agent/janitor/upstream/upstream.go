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

// Type & Method Definitions ...
//
type UpsManager struct {
	Upstreams []*Upstream `json:"upstreams"`
	sync.RWMutex
}

type Upstream struct {
	Name     string     `json:"name"`     // uniq name
	Alias    string     `json:"alias"`    // advertised url
	Listen   string     `json:"listen"`   // listen addr
	Target   string     `json:"target"`   // target addr
	Sticky   bool       `json:"sticky"`   // session sticky enabled (default no)
	Backends []*Backend `json:"backends"` // backend servers

	sessions *Sessions // runtime
	balancer Balancer  // runtime
}

func (u *Upstream) String() string {
	return fmt.Sprintf("name=%s, alias=%s, listen=%s, sticky=%v", u.Name, u.Alias, u.Listen, u.Sticky)
}

func newUpstream(first *BackendCombined) *Upstream {
	return &Upstream{
		Name:     first.Upstream.Name,
		Alias:    first.Upstream.Alias,
		Listen:   first.Upstream.Listen,
		Target:   first.Upstream.Target,
		Sticky:   first.Upstream.Sticky,
		Backends: []*Backend{first.Backend},
		sessions: newSessions(), // sessions store
		balancer: &wrrBalancer{
			index: -1,
			cw:    0,
		}, // default balancer
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
		if v.ID == name || v.CleanName == name {
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
	ID         string  `json:"id"`          // backend server id(name)
	IP         string  `json:"ip"`          // backend server ip
	Port       uint64  `json:"port"`        // backend server port
	TargetPort uint64  `json:"target_port"` // target port
	Scheme     string  `json:"scheme"`      // http / https, auto detect & setup by httpProxy
	Version    string  `json:"version"`
	Weight     float64 `json:"weihgt"`
	CleanName  string  `json:"clean_name"` // backend server clean id(name)
}

func (b *Backend) String() string {
	return fmt.Sprintf("id=%s, addr=%s, weight=%.2f", b.ID, b.Addr(), b.Weight)
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

// BackendCombined
type BackendCombined struct {
	*Upstream `json:"upstream"`
	*Backend  `json:"backend"`
}

func (cmb *BackendCombined) String() string {
	return fmt.Sprintf("upstream: [%s], backend: [%s]", cmb.Upstream, cmb.Backend)
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
	if !strings.HasSuffix(cmb.Backend.ID, "."+cmb.Upstream.Name) {
		return errors.New("backend name must be suffixed by upstream name")
	}
	return nil
}

func (cmb *BackendCombined) Format() {
	// rewrite Upstream.Listen
	cmb.Upstream.Listen = cmb.Upstream.tcpListen()

	// rewrite backend clean name
	fields := strings.SplitN(cmb.Backend.ID, ".", 2)
	if len(fields) == 2 {
		cmb.Backend.CleanName = fields[1]
	}
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
		name    = cmb.Upstream.Name
		alias   = cmb.Upstream.Alias
		listen  = cmb.Upstream.Listen
		target  = cmb.Upstream.Target
		backend = cmb.Backend.ID
	)

	_, u := getUpstreamByNameAndTarget(name, target)
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
	u.Sticky = cmb.Upstream.Sticky

	// update backend
	b.IP = cmb.Backend.IP
	b.Port = cmb.Backend.Port
	b.Scheme = cmb.Backend.Scheme
	b.Version = cmb.Backend.Version
	b.Weight = cmb.Backend.Weight

	return
}

func GetBackend(u *Upstream, backend string) *Backend {
	mgr.RLock()
	defer mgr.RUnlock()

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

	return Lookup(remoteIP, u, "")
}

// similar as lookup, but by upstream listen
func LookupListen(remoteIP, listen string) *BackendCombined {
	mgr.RLock()
	_, u := getUpstreamByListen(listen)
	mgr.RUnlock()

	if u == nil {
		return nil
	}

	return Lookup(remoteIP, u, "")
}

func LookupUpstream(remoteIP, name, port, backend string) *BackendCombined {
	var up *Upstream
	mgr.RLock()
	for _, u := range mgr.Upstreams {
		if u.Name == name && u.Target == port {
			up = u
		}
	}
	mgr.RUnlock()

	if up == nil {
		return nil
	}

	return Lookup(remoteIP, up, backend)
}

// lookup select a suitable backend according by sessions & balancer
func Lookup(remoteIP string, u *Upstream, backend string) *BackendCombined {
	var b *Backend

	if u == nil {
		return nil
	}

	defer func() {
		if u.Sticky && b != nil {
			u.sessions.update(remoteIP, b)
		}
	}()

	// obtain specified backend
	if backend != "" {
		b = GetBackend(u, backend)
		if b == nil {
			return nil
		}
		return &BackendCombined{u, b}
	}

	// obtain session by remoteIP
	if u.Sticky {
		if b = u.sessions.get(remoteIP); b != nil {
			return &BackendCombined{u, b}
		}
	}

	// use balancer to obtain a new backend
	if b = nextBackend(u); b != nil {
		return &BackendCombined{u, b}
	}

	return nil
}

func nextBackend(u *Upstream) *Backend {
	mgr.RLock()
	defer mgr.RUnlock()

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

func getUpstreamByNameAndTarget(name, target string) (int, *Upstream) {
	for i, v := range mgr.Upstreams {
		if v.Name == name && v.Target == target {
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
