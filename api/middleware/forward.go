package middleware

import (
	"bufio"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Dataman-Cloud/swan/api"

	log "github.com/Sirupsen/logrus"
)

type forwardMiddleware struct {
	leader string
}

func NewForwardMiddleware(leader string) *forwardMiddleware {
	return &forwardMiddleware{
		leader: leader,
	}
}

func (m *forwardMiddleware) Name() string {
	return "forward"
}

func (m *forwardMiddleware) WrapHandler(handler api.HandlerFunc) api.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// NOTE(nmg): If you just use ip address here, the `url.Parse` with get error with
		// `first path segment in URL cannot contain colon`.
		// It's golang 1.8's bug. more details see https://github.com/golang/go/issues/18824.
		leaderUrl := m.leader
		if !strings.HasPrefix(m.leader, "http://") {
			leaderUrl = "http://" + m.leader
		}

		leaderURL, err := url.Parse(leaderUrl + r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		rr, err := http.NewRequest(r.Method, leaderURL.String(), r.Body)
		rr.URL.RawQuery = r.URL.RawQuery
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}

		copyHeader(r.Header, &rr.Header)

		// Create a client and query the target
		client := &http.Client{}
		lresp, err := client.Do(rr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		log.Infof("Request forwarding %s %s %s", rr.Method, rr.URL, lresp.Status)

		dH := w.Header()
		copyHeader(lresp.Header, &dH)
		dH.Add("Requested-Host", rr.Host)

		reader := bufio.NewReader(lresp.Body)
		for {
			line, err := reader.ReadBytes('\n')

			if err == io.EOF {
				if _, err := w.Write(line); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return err
				}

				return nil
			}

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}

			if len(line) == 0 {
				continue
			}

			if _, err := w.Write(line); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return err
			}
		}

		return nil
	}
}

func copyHeader(src http.Header, dest *http.Header) {
	for n, v := range src {
		for _, vv := range v {
			dest.Set(n, vv)
		}
	}
}
