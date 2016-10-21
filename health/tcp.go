package health

import (
	"net"
	"time"

	"github.com/Sirupsen/logrus"
)

type TCPChecker struct {
	ID          string
	Addr        string
	Interval    int
	Timeout     int
	MaxFailures int

	FailedHandler HandlerFunc

	AppID  string
	TaskID string

	quit chan struct{}
}

func NewTCPChecker(id, addr string, port int, interval, timeout, failures int, handler HandlerFunc, appId, taskId string) *TCPChecker {
	return &TCPChecker{
		ID:            id,
		Addr:          addr,
		Interval:      interval,
		Timeout:       timeout,
		MaxFailures:   failures,
		FailedHandler: handler,
		AppID:         appId,
		TaskID:        taskId,
		quit:          make(chan struct{}),
	}
}

func (c *TCPChecker) Start() {
	ticker := time.NewTicker(time.Duration(c.Interval) * time.Second)
	maxFailures := 0
	for {

		if maxFailures >= c.MaxFailures {
			c.FailedHandler(c.AppID, c.TaskID)
			return
		}

		select {
		case <-ticker.C:
			_, err := net.ResolveTCPAddr("tcp", c.Addr)
			if err != nil {
				logrus.Errorf("Resolve tcp addr failed: %s", err.Error())
				return
			}

			if conn, err := net.DialTimeout("tcp",
				c.Addr,
				time.Duration(c.Timeout)*time.Second); err != nil {
				logrus.WithFields(logrus.Fields{"protocol": "tcp",
					"address":  c.Addr,
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
				"interval": c.Interval,
				"timeout":  c.Timeout},
			).Info("[OK] check service")

		case <-c.quit:
			return
		}
	}

}

func (c *TCPChecker) Stop() {
	close(c.quit)
}
