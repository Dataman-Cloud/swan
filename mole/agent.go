package mole

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	FILE_UUID = "/etc/.mole.uuid"

	errNotConnected = errors.New("not connected to master")
	errClosed       = errors.New("agent listener closed")
)

type Agent struct {
	id      string   // unique agent id
	master  *url.URL // master url
	backend *url.URL // backend url, TODO: support multi backends
	conn    net.Conn // control connection to master
	api     *agentApi

	mux    sync.RWMutex  // protect flag closed
	closed bool          // flag on pool closed
	pool   chan net.Conn // worker connection pool
}

func NewAgent(cfg *Config) *Agent {
	id, err := getAgentID()
	if err != nil {
		log.Fatalln(err)
	}

	a := &Agent{
		id:      id,
		master:  cfg.master,
		backend: cfg.backend,
		pool:    make(chan net.Conn, 1024),
	}
	a.api = newAgentApi(a.serveProxy)
	return a
}

func getAgentID() (string, error) {
	_, err := os.Stat(FILE_UUID)
	if os.IsNotExist(err) {
		uuid := randNumber(16)
		err = ioutil.WriteFile(FILE_UUID, []byte(uuid), os.FileMode(0400))
		return string(uuid), err
	}

	bs, err := ioutil.ReadFile(FILE_UUID)
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(bs)), nil
}

func (a *Agent) Join() error {
	conn, err := net.DialTimeout("tcp", a.master.Host, time.Second*10)
	if err != nil {
		return fmt.Errorf("agent Join error: %v", err)
	}

	// Disable IO Read TimeOut
	conn.SetReadDeadline(time.Time{})
	// Setting TCP KeepAlive on the socket connection will prohibit
	// ECONNTIMEOUT unless the socket connection truly is broken
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}
	a.conn = conn

	// send join cmd
	command := newCmd(cmdJoin, a.id, "")
	_, err = conn.Write(command)

	return err
}

func (a *Agent) Serve() error {
	if a.conn == nil {
		return errNotConnected
	}
	defer a.conn.Close()

	var (
		errChProt = make(chan error)
		errChHTTP = make(chan error)
	)
	go func() {
		errChProt <- a.ServeProtocol() // serve cluster protocol
	}()
	go func() {
		errChHTTP <- a.ServeApis() // serve http api
	}()

	select {
	case err := <-errChProt:
		return fmt.Errorf("protocol serving error: %v", err)
	case err := <-errChHTTP:
		return fmt.Errorf("httpApi serving error: %v", err)
	}

	return errors.New("never be here")
}

func (a *Agent) ServeProtocol() error {
	// protocol decoder
	dec := NewDecoder(a.conn)

	for {
		cmd, err := dec.Decode()
		if err != nil { // control conn closed, exit Serve() to trigger agent ReJoin
			return fmt.Errorf("agent decode protocol error: %v", err)
		}
		if err := cmd.valid(); err != nil {
			log.Errorf("agent received invalid command: %v", err)
			continue
		}

		// handle master command
		switch cmd.Cmd {

		case cmdNewWorker: // launch a new tcp connection as the worker connection
			connWorker, err := net.DialTimeout("tcp", a.master.Host, time.Second*10)
			if err != nil {
				log.Errorf("agent dial master error: %v", err)
				continue
			}
			command := newCmd(cmdNewWorker, a.id, cmd.WorkerID)
			_, err = connWorker.Write(command)
			if err != nil {
				log.Errorf("agent notify back worker id error: %v", err)
				continue
			}

			// put the worker conn to agent connection pools
			go a.HandleWorkerConn(connWorker)

		case cmdPing:
			pong := newCmd(cmdPing, a.id, "")
			if _, err := a.conn.Write(pong); err != nil {
				log.Errorf("agent heart pong error: %v", err)
			}
		}
	}

	return nil
}

// ServeApis serve agent-local http or backend http services
func (a *Agent) ServeApis() error {
	a.api.SetupRoutes()
	server := &http.Server{
		Handler: a.api,
	}
	return server.Serve(a)
}

func (a *Agent) serveProxy(w http.ResponseWriter, r *http.Request) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		w.WriteHeader(500)
		return
	}

	connMaster, _, err := hijacker.Hijack()
	if err != nil {
		w.WriteHeader(500)
		return
	}
	defer connMaster.Close()

	connBackend, err := a.dialBackend()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer connBackend.Close()

	go func() {
		r.Write(connBackend)
	}()

	io.Copy(connMaster, connBackend)
}

func (a *Agent) dialBackend() (net.Conn, error) {
	switch a.backend.Scheme {
	case "unix":
		return net.Dial(a.backend.Scheme, a.backend.Path)
	case "tcp", "http", "https":
		return net.Dial("tcp", a.backend.Host)
	}
	return nil, errors.New("not supported backend scheme")
}

// put the worker connection to the pool
func (a *Agent) HandleWorkerConn(conn net.Conn) error {
	a.mux.RLock()
	defer a.mux.RUnlock()
	if a.closed {
		return errClosed
	}
	a.pool <- conn
	return nil
}

// implement net.Listener interface (process the cached worker connection in the pool)
func (a *Agent) Accept() (net.Conn, error) {
	conn, ok := <-a.pool
	if !ok {
		return nil, errClosed
	}
	return conn, nil
}

func (a *Agent) Close() error {
	a.mux.Lock()
	a.closed = true
	close(a.pool)
	a.mux.Unlock()
	return nil
}

func (a *Agent) Addr() net.Addr {
	return a.conn.LocalAddr()
}

//
// agentApi is a simple http mutex router
type agentApi struct {
	m               map[string]map[string]http.HandlerFunc // method -> path -> handleFunc
	notFoundHandler http.HandlerFunc
}

func newAgentApi(notFoundHandler http.HandlerFunc) *agentApi {
	return &agentApi{
		m:               make(map[string]map[string]http.HandlerFunc),
		notFoundHandler: notFoundHandler,
	}
}

func (api *agentApi) SetupRoutes() {
	api.m = map[string]map[string]http.HandlerFunc{
		"GET": map[string]http.HandlerFunc{
			"/hello": api.hello,
		},
	}
}

// implement http.Handler interface
func (api *agentApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		method = r.Method
		path   = r.URL.Path
	)

	if m, ok := api.m[method]; ok {
		if h, ok := m[path]; ok {
			h(w, r)
			return
		}
	}

	if h := api.notFoundHandler; h != nil {
		h(w, r)
		return
	}

	http.Error(w, "page not found", 404)
}

func (api *agentApi) hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("hello world"))
}
