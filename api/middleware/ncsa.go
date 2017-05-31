package middleware

import (
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/Dataman-Cloud/swan/api"
)

type NCSACommonLogMiddleware struct {
}

func NewNCSACommonLogMiddleware() *NCSACommonLogMiddleware {
	return &NCSACommonLogMiddleware{}
}

func (m *NCSACommonLogMiddleware) Name() string {
	return "ncsa"
}

func (m *NCSACommonLogMiddleware) WrapHandler(handler api.HandlerFunc) api.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var username = "-"
		if r.URL.User != nil {
			if name := r.URL.User.Username(); name != "" {
				username = name
			}
		}
		err := handler(w, r)
		log.Printf("%s - %s [%s] \"%s %s %s\" %d %d",
			strings.Split(r.RemoteAddr, ":")[0],
			username,
			time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			r.URL.RequestURI(),
			r.Proto,
			//w.StatusCode,
			//w.ContentLength,
		)

		return err
	}
}
