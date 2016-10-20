package health

import (
	"net"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

type HTTPChecker struct {
	Url         string
	Interval    int
	Timeout     int
	MaxFailures int

	FailedHandler func() error
}

func NewHttpChecker(url string, interval, timeout int) *HTTPChecker {
	return &HTTPChecker{
		Url:      url,
		Interval: interval,
		Timeout:  timeout,
	}
}

func (c *HTTPChecker) Start() {
	go func() {
		ticker := time.NewTicker(time.Duration(c.Interval) * time.Second)

		maxFailures := 0

		for {

			if maxFailures >= c.MaxFailures {
				c.FailedHandler()
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

			}
		}

	}()
}
