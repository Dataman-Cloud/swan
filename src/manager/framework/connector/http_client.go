package connector

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"time"
)

const (
	HTTP_TIMEOUT_DURATION   = 10 * time.Second
	HTTP_KEEPALIVE_DURATION = 30 * time.Second

	USER_AGENT       = "swan/0.1"
	MESOS_STREAM_KEY = "swan/0.1"
)

type HttpClient struct {
	streamID string
	url      string
	client   *http.Client
}

func NewHTTPClient(addr, path string) *HttpClient {
	return &HttpClient{
		url: "http://" + addr + path,
		client: &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   HTTP_TIMEOUT_DURATION,
					KeepAlive: HTTP_KEEPALIVE_DURATION,
				}).Dial,
			},
		},
	}
}

func (c *HttpClient) send(payload []byte) (*http.Response, error) {
	httpReq, err := http.NewRequest("POST", c.url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", USER_AGENT)
	if c.streamID != "" {
		httpReq.Header.Set("Mesos-Stream-Id", c.streamID)
	}

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Unable to do request: %s", err)
	}

	if httpResp.Header.Get("Mesos-Stream-Id") != "" {
		c.streamID = httpResp.Header.Get("Mesos-Stream-Id")
	}

	return httpResp, nil
}
