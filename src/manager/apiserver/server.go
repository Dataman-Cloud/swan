package apiserver

import (
	"net"
	"net/http"
	"sync"

	"github.com/Dataman-Cloud/swan/src/manager/apiserver/router"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type ApiServer struct {
	addr    string
	sock    string
	routers []router.Router
}

func NewApiServer(addr, sock string) *ApiServer {
	return &ApiServer{
		addr: addr,
		sock: sock,
	}
}

// createMux initializes the main router the server uses.
func (s *ApiServer) createMux() *mux.Router {
	m := mux.NewRouter()

	logrus.Debug("Registering routers")
	for _, router := range s.routers {
		for _, r := range router.Routes() {
			f := s.makeHTTPHandler(r.Handler())

			logrus.Debugf("Registering %s, %s", r.Method(), r.Path())
			m.Path(r.Path()).Methods(r.Method()).Handler(f)
		}
	}

	return m
}

func (s *ApiServer) makeHTTPHandler(handler router.APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{"from": r.RemoteAddr}).Infof("[%s] %s", r.Method, r.URL.Path)
		if err := handler(w, r); err != nil {
			logrus.Errorf("Handler for %s %s returned error: %v", r.Method, r.URL.Path, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// InitRouter initializes the list of routers for the server.
func (s *ApiServer) AppendRouter(routers ...router.Router) {
	for _, r := range routers {
		s.routers = append(s.routers, r)
	}
}

func (s *ApiServer) ListenAndServe() error {
	var wg sync.WaitGroup
	var e error

	wg.Add(2)

	go func() {
		srv := &http.Server{
			Addr:    s.addr,
			Handler: s.createMux(),
		}
		logrus.Infof("API Server listen on %s", s.addr)
		ln, err := net.Listen("tcp", s.addr)
		if err != nil {
			logrus.Errorf("Listen on %s error: %s", s.addr, err)
			e = err
		}
		wg.Done()
		srv.Serve(ln)
	}()

	go func() {
		srv := &http.Server{
			Addr:    s.sock,
			Handler: s.createMux(),
		}
		logrus.Infof("API Server listen on %s", s.sock)
		ln, err := net.ListenUnix("unix", &net.UnixAddr{
			Name: s.sock,
			Net:  "unix",
		})
		if err != nil {
			logrus.Errorf("Listen on %s error: %s", s.sock, err)
			e = err
		}

		wg.Done()
		srv.Serve(ln)
	}()

	wg.Wait()
	return e
}
