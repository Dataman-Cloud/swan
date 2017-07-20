package api

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

func (r *Server) enableDebug(w http.ResponseWriter, req *http.Request) {
	logrus.SetLevel(logrus.DebugLevel)   // force set to debug level
	writeJSON(w, 200, map[string]string{ // tell previous log level
		"previous": r.cfg.LogLevel,
	})
}

func (r *Server) disableDebug(w http.ResponseWriter, req *http.Request) {
	l, err := logrus.ParseLevel(r.cfg.LogLevel)
	if err != nil {
		l = logrus.InfoLevel
	}
	logrus.SetLevel(l) // restore to previous log level
}
