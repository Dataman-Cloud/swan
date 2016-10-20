package health

import (
	"fmt"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
)

type TCPChecker struct {
	Addr        string
	Port        int
	Interval    int
	Timeout     int
	MaxFailures int

	FailedHandler func() error
}

func NewTcpChecker(addr string, port int, interval, timeout, failures int) *TCPChecker {
	return &TCPChecker{
		Addr:        addr,
		Port:        port,
		Interval:    interval,
		Timeout:     timeout,
		MaxFailures: failures,
	}
}

func (c *TCPChecker) Start() {
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
				if conn, err := net.DialTimeout("tcp",
					fmt.Sprintf("%s:%d", c.Addr, c.Port),
					time.Duration(c.Timeout)*time.Second); err != nil {
					logrus.WithFields(logrus.Fields{"protocol": "tcp",
						"address":  c.Addr,
						"port":     c.Port,
						"interval": c.Interval,
						"timeout":  c.Timeout},
					).Error("[FAILED] check service")

					if conn != nil {
						conn.Close()
					}

					maxFailures += 1
					break
				}

				logrus.WithFields(logrus.Fields{"protocol": "tcp",
					"address":  c.Addr,
					"port":     c.Port,
					"interval": c.Interval,
					"timeout":  c.Timeout},
				).Info("[OK] check service")

			}
		}

	}()
}
