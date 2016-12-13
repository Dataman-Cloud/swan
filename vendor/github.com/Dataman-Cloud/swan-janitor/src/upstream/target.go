package upstream

import (
	"fmt"
	"net"
	"net/url"

	log "github.com/Sirupsen/logrus"
)

type Target struct {
	Node           string
	Address        string
	ServiceName    string
	ServiceID      string
	ServiceAddress string
	ServicePort    string
	Upstream       *Upstream
}

func (t *Target) Equal(t1 *Target) bool {
	return t.Node == t1.Node &&
		t.Address == t1.Address &&
		t.ServiceName == t1.ServiceName &&
		t.ServiceID == t1.ServiceID &&
		t.ServiceAddress == t1.ServiceAddress &&
		t.ServicePort == t1.ServicePort
}

func (t *Target) ToString() string {
	return fmt.Sprintf("%s-%s-%s-%s-%s-%s", t.Node, t.Address, t.ServiceName, t.ServiceID, t.ServiceAddress, t.ServicePort)
}

func (t Target) Entry() *url.URL {
	url, err := url.Parse(fmt.Sprintf("%s://%s", t.Upstream.FrontendProto, net.JoinHostPort(t.ServiceAddress, t.ServicePort)))
	if err != nil {
		log.Error("parse target.ServiceAddress %s to url got err %s", t.ServiceAddress, err)
	}

	return url
}
