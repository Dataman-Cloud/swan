package api

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"sync"
	"time"

	"github.com/Dataman-Cloud/swan/store"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type Config struct {
	Listen   string
	LogLevel string
}

type Server struct {
	cfg      *Config
	listener net.Listener // specified net listener
	leader   string
	server   *http.Server
	driver   Driver
	db       store.Store

	sync.Mutex
}

func NewServer(cfg *Config, l net.Listener, driver Driver, db store.Store) *Server {
	s := &Server{
		cfg:      cfg,
		listener: l,
		leader:   "",
		driver:   driver,
		db:       db,
	}

	s.server = &http.Server{
		Handler: s.createMux(),
	}

	return s
}

// createMux initializes the main router the server uses.
func (s *Server) createMux() *mux.Router {
	m := mux.NewRouter()

	s.setupRoutes(m)

	if s.cfg.LogLevel == "debug" {
		profilerSetup(m, "/debug/")
	}

	return m
}

func (s *Server) enableCORS(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, X-Registry-Auth")
	w.Header().Add("Access-Control-Allow-Methods", "HEAD, GET, POST, DELETE, PUT, OPTIONS")
}

func profilerSetup(r *mux.Router, path string) {
	var m = r.PathPrefix(path).Subrouter()
	m.HandleFunc("/pprof/", pprof.Index)
	m.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	m.HandleFunc("/pprof/profile", pprof.Profile)
	m.HandleFunc("/pprof/symbol", pprof.Symbol)
	m.HandleFunc("/pprof/trace", pprof.Trace)
	m.HandleFunc("/pprof/block", pprof.Handler("block").ServeHTTP)
	m.HandleFunc("/pprof/heap", pprof.Handler("heap").ServeHTTP)
	m.HandleFunc("/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	m.HandleFunc("/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
}

func (s *Server) makeHTTPHandler(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.enableCORS(w)

		if s.cfg.Listen != s.GetLeader() {
			if r.Method != "GET" {
				s.forwardRequest(w, r)
				return
			}
			handler(w, r)
			return
		}

		handler(w, r)
	}
}

func (s *Server) Run() error {
	return s.server.Serve(s.listener)
}

// gracefully shutdown.
func (s *Server) Shutdown() error {
	// If s.server is nil, api server is not running.
	if s.server != nil {
		// NOTE(nmg): need golang 1.8+ to run this method.
		return s.server.Shutdown(nil)
	}

	return nil
}

func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}

	return nil
}

func (s *Server) Reload() error {
	log.Println("Reload api server for leader change")

	if err := s.Stop(); err != nil {
		return fmt.Errorf("Shutdown api server error: %v", err.Error())
	}
	// NOTE(nmg): Sometimes the api server can't be closed immediately.
	// In this situation the `bind: address already in use` error will be occurred.
	// So we use a `for loop` to aviod this.
	// TODO(nmg): Fix this more elegant.
	for {
		err := s.Run()
		if strings.Contains(err.Error(), "bind: address already in use") {
			log.Errorf("Start apiserver error %s. Retry after 1 second.", err.Error())
			time.Sleep(1 * time.Second)
			continue
		}

		return fmt.Errorf("apiserver run error: %v", err)
	}

}

func (s *Server) UpdateLeader(leader string) {
	s.Lock()
	defer s.Unlock()

	s.leader = leader
}

func (s *Server) GetLeader() string {
	s.Lock()
	defer s.Unlock()

	return s.leader
}

func (s *Server) forwardRequest(w http.ResponseWriter, r *http.Request) {
	// dial leader
	dst, err := net.DialTimeout("tcp", s.leader, time.Second*60)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	err = r.WriteProxy(dst) // send original request
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Request forwarding %s %s --> %s", r.Method, r.URL, s.leader)

	// obtian the client underlying net.Conn
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, fmt.Sprintf("not support http hijack: %T", w), 500)
		return
	}

	src, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer src.Close()

	// io copy between src & dst
	errc := make(chan error, 2)
	cp := func(w io.WriteCloser, r io.Reader) {
		defer w.Close()
		_, err := io.Copy(w, r) // TODO caculate each piece of io buffer by real time
		errc <- err
	}

	go cp(dst, src)
	cp(src, dst) // note: hanging wait while copying the response

	err = <-errc
	if err != nil && err != io.EOF {
		err = fmt.Errorf("io copy error: %v", err)
		src.Write([]byte("HTTP/1.0 500 Internal Server Error\r\n\r\n" + err.Error() + "\r\n"))
		return
	}
}
