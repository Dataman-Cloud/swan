package ns

import (
	"net"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func TestFormatA(t *testing.T) {
	res := Resolver{}
	r, e := res.formatA("domain", "192.168.1.1")
	assert.Nil(t, e)
	assert.Equal(t, "domain", r.Hdr.Name)

	_, e1 := res.formatA("domain", "192.168.1.fooabr")
	assert.NotNil(t, e1)
}

func TestFormatSRV(t *testing.T) {
	res := Resolver{}
	r, e := res.formatSRV("domain", "192.168.1.1:1023")
	assert.Nil(t, e)
	assert.Equal(t, "domain", r.Hdr.Name)

	_, e1 := res.formatSRV("domain", "192.168.1.fooabr")
	assert.NotNil(t, e1)
}

func TestFormatSOA(t *testing.T) {
	res := Resolver{}
	r := res.formatSOA("192.168.1.1")
	assert.Equal(t, "192.168.1.1", r.Hdr.Name)
}

func TestFormatNS(t *testing.T) {
	res := Resolver{}
	r := res.formatNS("192.168.1.1")
	assert.Equal(t, "192.168.1.1", r.Hdr.Name)
}

func TestRcode(t *testing.T) {
	err := &ForwardError{}
	assert.Equal(t, dns.RcodeRefused, rcode(err))
}

type fakeResponseWriter struct {
	remoteAddr net.Addr
}

func (f fakeResponseWriter) LocalAddr() net.Addr {
	return nil
}
func (f fakeResponseWriter) RemoteAddr() net.Addr      { return f.remoteAddr }
func (f fakeResponseWriter) WriteMsg(*dns.Msg) error   { return nil }
func (f fakeResponseWriter) Write([]byte) (int, error) { return 0, nil }
func (f fakeResponseWriter) Close() error              { return nil }
func (f fakeResponseWriter) TsigStatus() error         { return nil }
func (f fakeResponseWriter) TsigTimersOnly(bool)       {}
func (f fakeResponseWriter) Hijack()                   {}

func TestIsUDP(t *testing.T) {
	responseWriter := fakeResponseWriter{remoteAddr: &net.UDPAddr{}}
	assert.Equal(t, true, isUDP(responseWriter))

	responseWriter1 := fakeResponseWriter{remoteAddr: &net.TCPAddr{}}
	assert.Equal(t, false, isUDP(responseWriter1))
}
