package ipam

import "github.com/Dataman-Cloud/swan/src/manager/swancontext"

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

// TODO(chunmingxu) if IpamAdapter is not a server may call it register
func (ipamAdapter *IpamAdapter) Start() error {
	return nil
}

func (ipamAdapter *IpamAdapter) Stop() error {
	return nil
}

func (ipamAdapter *IpamAdapter) Run() error {
	return nil
}
