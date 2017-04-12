package connector

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHttpClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print(w, "Hello, client")
	}))
	defer ts.Close()

	urlParsed, _ := url.Parse(ts.URL)

	assert.NotNil(t, urlParsed)
	client := NewHTTPClient("x", "y")
	assert.NotNil(t, client)
}

func TestSendPayload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get("Content-Type"), "application/x-protobuf")
		assert.Equal(t, r.Header.Get("User-Agent"), USER_AGENT)

		w.Header().Set("Mesos-Stream-Id", "foobar")
	}))
	defer ts.Close()

	urlParsed, _ := url.Parse(ts.URL)

	assert.NotNil(t, urlParsed)
	client := NewHTTPClient(urlParsed.Host, urlParsed.RawQuery)
	assert.NotNil(t, client)
	resp, err := client.send([]byte("foobar"))
	assert.NotNil(t, resp)
	assert.Nil(t, err)

	assert.Equal(t, resp.Header.Get("Mesos-Stream-Id"), "foobar")

}
