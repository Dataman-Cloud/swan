package janitor

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

// generic tcp proxy server
type tcpProxyServer struct {
	listenAddr string
	upstreams  *Upstreams
	stats      *Stats

	listener net.Listener

	sync.RWMutex // protect clients
	clients      map[string]net.Conn

	startedAt time.Time
	serving   bool
}

func (s *JanitorServer) newTCPProxyServer(listen string) *tcpProxyServer {
	return &tcpProxyServer{
		listenAddr: listen,
		upstreams:  s.upstreams,
		stats:      s.stats,
		clients:    make(map[string]net.Conn),
	}
}

func (p *tcpProxyServer) MarshalJSON() ([]byte, error) {
	p.RLock()
	n := len(p.clients)
	p.RUnlock()

	m := map[string]interface{}{
		"uptime":         time.Now().Sub(p.startedAt).String(),
		"listen":         p.listenAddr,
		"serving":        p.serving,
		"active_clients": n,
	}
	return json.Marshal(m)
}

func (p *tcpProxyServer) listen() error {
	p.startedAt = time.Now()

	l, err := net.Listen("tcp", p.listenAddr)
	if err != nil {
		return err
	}
	p.listener = l

	return nil
}

func (p *tcpProxyServer) serve() {
	defer func() {
		p.serving = false
	}()

	p.serving = true
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			log.Errorf("[TCP] listener :%s Accept error: %v", p.listen, err)
			return
		}

		go p.serveTCP(conn)
	}
}

func (p *tcpProxyServer) stop() {
	if p.listener != nil {
		p.listener.Close()
	}

	p.RLock()
	for _, conn := range p.clients {
		conn.Close()
	}
	p.RUnlock()
}

func (p *tcpProxyServer) lookup(conn net.Conn) (*Target, error) {
	var (
		local  = conn.LocalAddr().String()
		remote = conn.RemoteAddr().String()
	)

	_, localPort, err := net.SplitHostPort(local)
	if err != nil {
		return nil, err
	}
	remoteHost, _, err := net.SplitHostPort(remote)
	if err != nil {
		return nil, err
	}

	listen := ":" + localPort // TODO

	selected := p.upstreams.lookupListen(remoteHost, listen)
	if selected == nil {
		return nil, fmt.Errorf("no matched targets for request [%s]", listen)
	}

	log.Debugf("[TCP]: proxy redirecting request [%s] -> [%s] -> [%s-%s]",
		remoteHost, listen, selected.TaskID, selected.addr(),
	)
	return selected, nil
}

func (p *tcpProxyServer) serveTCP(conn net.Conn) {
	remote := conn.RemoteAddr().String()

	p.Lock()
	p.clients[remote] = conn
	p.Unlock()

	defer func() {
		conn.Close()

		p.Lock()
		delete(p.clients, remote)
		p.Unlock()
	}()

	var (
		err  error
		in   int64
		out  int64
		dGlb *deltaGlb // delta global
	)

	defer func() {
		if err != nil {
			conn.Write([]byte(err.Error()))
			log.Errorf("[TCP] proxy serve error: %v, recived:%d, transmitted:%d", err, in, out)
			dGlb = &deltaGlb{uint64(in), uint64(out), 1, 1}
		} else {
			log.Printf("[TCP] proxy serve succeed: recived:%d, transmitted:%d", in, out)
			dGlb = &deltaGlb{uint64(in), uint64(out), 1, 0}
		}
		p.stats.incr(nil, dGlb)
	}()

	// lookup a proper backend according by src ip & dest port
	selected, err := p.lookup(conn)
	if err != nil {
		return
	}

	var (
		addr = selected.addr()
		aid  = selected.AppID
		tid  = selected.TaskID
	)

	// do proxy
	p.stats.incr(&deltaApp{aid, tid, 1, 0, 0, 1}, nil) // conn, active
	in, out, err = p.doRawProxy(conn, addr)
	p.stats.incr(&deltaApp{aid, tid, -1, uint64(in), uint64(out), 0}, nil) // disconnect
}

func (p *tcpProxyServer) doRawProxy(src net.Conn, addr string) (int64, int64, error) {
	var in, out int64

	// dial backend
	dst, err := net.DialTimeout("tcp", addr, time.Second*60)
	if err != nil {
		err = fmt.Errorf("cannot connect to upstream %s: %v", addr, err)
		return in, out, err
	}
	defer dst.Close()

	// io copy between src & dst
	errc := make(chan error, 2)
	cp := func(w io.WriteCloser, r io.Reader, c *int64) {
		defer w.Close()

		n, err := io.Copy(w, r) // TODO caculate each piece of io buffer by real time
		if n > 0 {
			*c += n
		}
		errc <- err
	}

	go cp(dst, src, &in)
	cp(src, dst, &out) // note: hanging wait while copying the response

	err = <-errc
	if err != nil && err != io.EOF {
		err = fmt.Errorf("io copy error: %v", err)
		return in, out, err
	}
	return in, out, nil
}
