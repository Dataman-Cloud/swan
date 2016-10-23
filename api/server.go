package api

import (
	"net"
	"net/http"

	"github.com/Dataman-Cloud/swan/scheduler"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type Server struct {
	router *Router
}

func NewServer(sched *scheduler.Scheduler) *Server {
	return &Server{
		router: NewRouter(sched),
	}
}

func (s *Server) Routes() []Route {
	return s.router.routes
}

// createMux initializes the main router the server uses.
func (s *Server) createMux() *mux.Router {
	m := mux.NewRouter()

	logrus.Debug("Registering routers")
	for _, r := range s.Routes() {
		f := s.makeHTTPHandler(r.Handler())

		logrus.Debugf("Registering %s, %s", r.Method(), r.Path())
		m.Path(r.Path()).Methods(r.Method()).Handler(f)
	}

	return m
}

func (s *Server) makeHTTPHandler(handler APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{"from": r.RemoteAddr}).Infof("[%s] %s", r.Method, r.URL.Path)
		if err := handler(w, r); err != nil {
			logrus.Errorf("Handler for %s %s returned error: %v", r.Method, r.URL.Path, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: s.createMux(),
	}
	logrus.Infof("API Server listen on %s", addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		logrus.Errorf("Listen on %s error: %s", addr, err)
		return err
	}
	return srv.Serve(ln)
}
