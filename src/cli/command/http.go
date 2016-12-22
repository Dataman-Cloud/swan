package command

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
)

type Client struct {
	url    string
	client *http.Client
}

const APIVERSION = "/v_beta"

func UnixDialer(proto, sock string) (net.Conn, error) {
	return net.Dial("unix", "/var/run/swan.sock")
}

func NewHTTPClient(path string) *Client {
	return &Client{
		url: "http://unix.sock" + APIVERSION + path,
		client: &http.Client{
			Transport: &http.Transport{
				Dial: UnixDialer,
			},
		},
	}
}

func (c *Client) Post(payload []byte) (*http.Response, error) {
	httpReq, err := http.NewRequest("POST", c.url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "swan/0.1")

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Unable to do request: %s", err)
	}

	return httpResp, nil
}

func (c *Client) Delete() (*http.Response, error) {
	httpReq, err := http.NewRequest("DELETE", c.url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("User-Agent", "swan/0.1")
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Unable to do request: %s", err)
	}

	return httpResp, nil
}

func (c *Client) Get() (*http.Response, error) {
	httpReq, err := http.NewRequest("GET", c.url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("User-Agent", "swan/0.1")
	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Unable to do request: %s", err)
	}

	return httpResp, nil
}

func (c *Client) Patch(payload []byte) (*http.Response, error) {
	httpReq, err := http.NewRequest("PATCH", c.url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "swan/0.1")

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Unable to do request: %s", err)
	}

	return httpResp, nil
}

func (c *Client) Put(payload []byte) (*http.Response, error) {
	httpReq, err := http.NewRequest("PUT", c.url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "swan/0.1")

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Unable to do request: %s", err)
	}

	return httpResp, nil
}
