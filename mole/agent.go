package mole

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	FILE_UUID = "/etc/.mole.uuid"

	errNotConnected = errors.New("not connected to master")
	errClosed       = errors.New("agent listener closed")
)

type ConnHandler interface {
	HandleWorkerConn(c net.Conn) error
}

type Agent struct {
	id      string      // unique agent id
	master  *url.URL    // master url
	conn    net.Conn    // control connection to master
	handler ConnHandler // worker connection handler
}

func NewAgent(cfg *Config) *Agent {
	id, err := getAgentID()
	if err != nil {
		log.Fatalln(err)
	}

	return &Agent{
		id:     id,
		master: cfg.Master,
	}
}

func getAgentID() (string, error) {
	_, err := os.Stat(FILE_UUID)
	if os.IsNotExist(err) {
		uuid := randNumber(16)
		err = ioutil.WriteFile(FILE_UUID, []byte(uuid), os.FileMode(0400))
		return string(uuid), err
	}

	bs, err := ioutil.ReadFile(FILE_UUID)
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(bs)), nil
}

func (a *Agent) Join() error {
	conn, err := net.DialTimeout("tcp", a.master.Host, time.Second*10)
	if err != nil {
		return fmt.Errorf("agent Join error: %v", err)
	}

	// Disable IO Read TimeOut
	conn.SetReadDeadline(time.Time{})
	// Setting TCP KeepAlive on the socket connection will prohibit
	// ECONNTIMEOUT unless the socket connection truly is broken
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}
	a.conn = conn

	// send join cmd
	command := newCmd(cmdJoin, a.id, "")
	_, err = conn.Write(command)
	return err
}

func (a *Agent) ServeProtocol() error {
	// ensure joined
	if a.conn == nil {
		return errNotConnected
	}
	defer a.conn.Close()

	// protocol decoder
	dec := NewDecoder(a.conn)

	for {
		cmd, err := dec.Decode()
		if err != nil {
			return fmt.Errorf("agent decode protocol error: %v", err) // control conn closed, exit Serve() to trigger agent ReJoin
		}
		if err := cmd.valid(); err != nil {
			log.Errorf("agent received invalid command: %v", err)
			continue
		}

		// handle master command
		switch cmd.Cmd {

		case cmdNewWorker: // launch a new tcp connection as the worker connection
			log.Debugln("agent launch a new tcp worker connection ...")

			connWorker, err := net.DialTimeout("tcp", a.master.Host, time.Second*10)
			if err != nil {
				log.Errorf("agent dial master error: %v", err)
				continue
			}
			command := newCmd(cmdNewWorker, a.id, cmd.WorkerID)
			_, err = connWorker.Write(command)
			if err != nil {
				log.Errorf("agent notify back worker id error: %v", err)
				continue
			}

			// put the worker conn to agent connection pools
			go a.handler.HandleWorkerConn(connWorker)

		case cmdPing:
			pong := newCmd(cmdPing, a.id, "")
			if _, err := a.conn.Write(pong); err != nil {
				log.Errorf("agent heart pong error: %v", err)
			}
		}
	}

	return nil
}

func (a *Agent) NewListener() net.Listener {
	l := &AgentListener{
		pool: make(chan net.Conn),
	}
	a.handler = l
	return l
}

type AgentListener struct {
	sync.RWMutex               // protect flag closed
	closed       bool          // flag on pool closed
	pool         chan net.Conn // worker connection pool
}

// implement net.Listener interface
// the caller could process the cached worker connection in the pool via `AgentListener`
func (l *AgentListener) Accept() (net.Conn, error) {
	conn, ok := <-l.pool
	if !ok {
		return nil, errClosed
	}
	return conn, nil
}

func (l *AgentListener) Close() error {
	l.Lock()
	if !l.closed {
		l.closed = true
		close(l.pool) // so the Accept() returned immediately
	}
	l.Unlock()
	return nil
}

func (l *AgentListener) Addr() net.Addr {
	return net.Addr(nil)
}

// implement ConnHandler interface
// put the worker connection to the pool
func (l *AgentListener) HandleWorkerConn(conn net.Conn) error {
	l.RLock()
	defer l.RUnlock()
	if l.closed {
		return errClosed
	}
	l.pool <- conn
	return nil
}
