package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func newHTTPProxy(t *url.URL, tr http.RoundTripper, flush time.Duration) http.Handler {
	rp := httputil.NewSingleHostReverseProxy(t)
	rp.Transport = tr
	rp.FlushInterval = flush
	rp.Transport = &meteredRoundTripper{tr}
	return rp
}

type meteredRoundTripper struct {
	tr http.RoundTripper
}

func (m *meteredRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	resp, err := m.tr.RoundTrip(r)
	return resp, err
}
