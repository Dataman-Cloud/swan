package api

import (
	"github.com/Dataman-Cloud/swan/api/router"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"net"
	"net/http"
)

type Server struct {
	addr    string
	routers []router.Router
}

func NewServer(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

// createMux initializes the main router the server uses.
func (s *Server) createMux() *mux.Router {
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

func (s *Server) makeHTTPHandler(handler router.APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{"from": r.RemoteAddr}).Infof("[%s] %s", r.Method, r.URL.Path)
		if err := handler(w, r); err != nil {
			logrus.Errorf("Handler for %s %s returned error: %v", r.Method, r.URL.Path, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// InitRouter initializes the list of routers for the server.
func (s *Server) InitRouter(routers ...router.Router) {
	for _, r := range routers {
		s.routers = append(s.routers, r)
	}
}

func (s *Server) ListenAndServe() error {
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.createMux(),
	}
	logrus.Infof("API Server listen on %s", s.addr)
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		logrus.Errorf("Listen on %s error: %s", s.addr, err)
		return err
	}
	return srv.Serve(ln)
}
