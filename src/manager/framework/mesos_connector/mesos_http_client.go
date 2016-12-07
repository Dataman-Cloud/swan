package mesos_connector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	mesosjson "github.com/Dataman-Cloud/swan/src/mesosproto/json"
)

type MesosHttpClient struct {
	StreamID string
	url      string
	client   *http.Client
}

func NewHTTPClient(addr, path string) *MesosHttpClient {
	return &MesosHttpClient{
		url: "http://" + addr + path,
		client: &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
			},
		},
	}
}

func (c *MesosHttpClient) Send(payload []byte) (*http.Response, error) {

	httpReq, err := http.NewRequest("POST", c.url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "swan/0.1")
	if c.StreamID != "" {
		httpReq.Header.Set("Mesos-Stream-Id", c.StreamID)
	}
	//log.Printf("SENDING:%v", httpReq)

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Unable to do request: %s", err)
	}
	if httpResp.Header.Get("Mesos-Stream-Id") != "" {
		c.StreamID = httpResp.Header.Get("Mesos-Stream-Id")
	}
	return httpResp, nil
}

func (c *MesosHttpClient) SendAsJson(call *mesosjson.Call) (*http.Response, error) {
	payload := new(bytes.Buffer)
	if err := json.NewEncoder(payload).Encode(call); err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.url, payload)
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
	c.StreamID = httpResp.Header.Get("Mesos-Stream-Id")
	log.Println("Stream-ID: ", c.StreamID)
	return httpResp, nil
}
