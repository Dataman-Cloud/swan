package listener

import (
	"log"
	"net"
	"net/http"
	//"time"

	"github.com/Dataman-Cloud/swan-janitor/src/config"

	"github.com/armon/go-proxyproto"
)

// http://www.hydrogen18.com/blog/stop-listening-http-server-go.html
// listener was not shutting down gracefully

func ListenAndServeHTTP(h http.Handler, ConfigProxy config.Proxy) {
	srv := &http.Server{
		Handler: h,
		Addr:    "",
	}

	if err := serve(srv); err != nil {
		log.Fatal("[FATAL] ", err)
	}
}

func serve(srv *http.Server) error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatal("[FATAL] ", err)
	}

	ln = &proxyproto.Listener{Listener: TcpKeepAliveListener{ln.(*net.TCPListener)}}

	return srv.Serve(ln)
}

// copied from http://golang.org/src/net/http/server.go?s=54604:54695#L1967
// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type TcpKeepAliveListener struct {
	*net.TCPListener
}

// TODO make this configurable to reduce TIME_WAIT issue
func (ln TcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	//if err = tc.SetKeepAlive(true); err != nil {
	//return
	//}
	//if err = tc.SetKeepAlivePeriod(3 * time.Minute); err != nil {
	//return
	//}
	return tc, nil
}
