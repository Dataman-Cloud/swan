package ipam

import (
	"github.com/Dataman-Cloud/swan/src/manager/swancontext"
)

type IpamAdapter struct {
	IPAM     *IPAM
	scontext *swancontext.SwanContext
}

func New(scontext *swancontext.SwanContext) (*IpamAdapter, error) {
	store, err := NewBoltStore(scontext.Config.IPAM.StorePath)
	if err != nil {
		return nil, err
	}

	m := NewIPAM(store)

	adapter := &IpamAdapter{
		scontext: scontext,
		IPAM:     m,
	}

	return adapter, nil
}

func (ipamAdapter *IpamAdapter) Start() error {
	ipamAdapter.scontext.ApiServer.AppendRouter(NewRouter(ipamAdapter.IPAM))
	return nil
}
func (ipamAdapter *IpamAdapter) Stop() error {
	return nil
}

func (ipamAdapter *IpamAdapter) Run() error {
	return nil
}
