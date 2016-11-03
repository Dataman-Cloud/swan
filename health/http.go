package health

import (
	"net"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

type HTTPChecker struct {
	ID          string
	Url         string
	Interval    int64
	Timeout     int64
	MaxFailures int64

	FailedHandler HandlerFunc

	AppID  string
	TaskID string

	quit chan struct{}
}

func NewHTTPChecker(id, url string, interval, timeout, failures int64, handler HandlerFunc, appId, taskId string) *HTTPChecker {
	return &HTTPChecker{
		ID:            id,
		Url:           url,
		Interval:      interval,
		Timeout:       timeout,
		MaxFailures:   failures,
		FailedHandler: handler,
		AppID:         appId,
		TaskID:        taskId,
		quit:          make(chan struct{}),
	}
}

func (c *HTTPChecker) Start() {
	ticker := time.NewTicker(time.Duration(c.Interval) * time.Second)

	maxFailures := 0

	for {

		if maxFailures >= int(c.MaxFailures) {
			c.FailedHandler(c.AppID, c.TaskID)
			return
		}

		select {
		case <-ticker.C:
			transport := http.Transport{
				Dial: func(network, addr string) (net.Conn, error) {
					return net.DialTimeout(network, addr, time.Duration(c.Timeout)*time.Second)
				},
			}

			client := http.Client{
				Transport: &transport,
			}

			resp, err := client.Head(c.Url)
			if err != nil {
				logrus.WithFields(logrus.Fields{"protocol": "http",
					"url":      c.Url,
					"interval": c.Interval,
					"timeout":  c.Timeout},
				).Error("[FAILED] check service")
				maxFailures += 1
				break
			}

			if resp.StatusCode != 200 {
				logrus.WithFields(logrus.Fields{"protocol": "http",
					"url":      c.Url,
					"interval": c.Interval,
					"timeout":  c.Timeout},
				).Error("[FAILED] check service")
				maxFailures += 1
				break
			}
			logrus.WithFields(logrus.Fields{"protocol": "http",
				"url":      c.Url,
				"interval": c.Interval,
				"timeout":  c.Timeout},
			).Info("[OK] check service")
		case <-c.quit:
			return
		}
	}

}

func (c *HTTPChecker) Stop() {
	close(c.quit)
}
