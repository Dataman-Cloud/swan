package manager

import (
	"bytes"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

type tcpMux struct {
	listen   string
	poolHTTP chan net.Conn
	poolMole chan net.Conn
}

func newTCPMux(l string) *tcpMux {
	return &tcpMux{
		listen:   l,
		poolHTTP: make(chan net.Conn, 1),
		poolMole: make(chan net.Conn, 1),
	}
}

func (m *tcpMux) ListenAndServe() error {
	l, err := net.Listen("tcp", m.listen)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Errorln("tcpMux Accept() error: %v", err)
			return err
		}

		go m.dispatch(conn)
	}
}

func (m *tcpMux) dispatch(conn net.Conn) {
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(time.Second * 30)
	}

	var (
		remote     = conn.RemoteAddr().String()
		header     = make([]byte, 4)
		headerCopy = bytes.NewBuffer(nil) // buffer to hold another copy of header
	)
	_, err := io.ReadFull(io.TeeReader(conn, headerCopy), header)
	if err != nil {
		log.Errorln("failed to read protocol header:", remote, err)
		return
	}

	bc := &bufConn{Conn: conn, reader: io.MultiReader(headerCopy, conn)}
	// dispatch to mole or http connection pool
	if bytes.Equal(header, []byte(`MOLE`)) {
		m.poolMole <- bc
		return
	}

	m.poolHTTP <- bc
}

func (m *tcpMux) NewHTTPListener(activate chan struct{}) net.Listener {
	return &muxListener{
		pool:     m.poolHTTP,
		activate: activate,
	}
}

func (m *tcpMux) NewMoleListener(activate chan struct{}) net.Listener {
	return &muxListener{
		pool:     m.poolMole,
		activate: activate,
	}
}

// implement net.Listener interface
// the caller could process the cached worker connection in the pool via `muxListener`
type muxListener struct {
	sync.Mutex               // protect flag closed
	closed     bool          // flag on pool closed
	pool       chan net.Conn // connection pool
	ready      bool
	activate   chan struct{}
}

func (l *muxListener) Accept() (net.Conn, error) {
	if l.ready {
		conn, ok := <-l.pool
		if !ok {
			return nil, errors.New("listener closed")
		}
		return conn, nil
	}

	<-l.activate
	l.ready = true
	return l.Accept()
}

func (l *muxListener) Close() error {
	l.Lock()
	if !l.closed {
		l.closed = true
		close(l.pool) // so the Accept() returned immediately
	}
	l.Unlock()
	return nil
}

func (l *muxListener) Addr() net.Addr {
	return net.Addr(nil)
}

// implement net.Conn with customized Read() method
type bufConn struct {
	net.Conn
	reader io.Reader
}

func (bc *bufConn) Read(bs []byte) (int, error) {
	return bc.reader.Read(bs)
}
