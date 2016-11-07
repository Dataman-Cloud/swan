package health

import (
	"github.com/Sirupsen/logrus"
	"net"
	"time"
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

			conn, err := net.DialTimeout("tcp",
				c.Addr,
				time.Duration(c.Timeout)*time.Second)
			if err != nil {
				logrus.Errorf("check task %s failed protocol %s address %s", c.TaskID, "tcp", c.Addr)

				maxFailures += 1
				break
			}

			conn.Close()

			//logrus.Infof("check task %s ok protocol %s address %s", c.TaskID, "tcp", c.Addr)

		case <-c.quit:
			return
		}
	}

}

func (c *TCPChecker) Stop() {
	close(c.quit)
}
