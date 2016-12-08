package state

import (
	"github.com/Dataman-Cloud/swan/src/mesosproto/mesos"
)

// wrapper offer to record offer reserve history
type OfferWrapper struct {
	Offer        *mesos.Offer
	CpusUsed     float64
	MemUsed      float64
	DiskUsed     float64
	PortUsedSize int
}

func NewOfferWrapper(offer *mesos.Offer) *OfferWrapper {
	o := &OfferWrapper{
		Offer:        offer,
		CpusUsed:     0,
		MemUsed:      0,
		DiskUsed:     0,
		PortUsedSize: 0,
	}
	return o
}

func (ow *OfferWrapper) PortsRemain() []uint64 {
	ports := make([]uint64, 0)
	for _, resource := range ow.Offer.Resources {
		if resource.GetName() == "ports" {
			for _, rang := range resource.GetRanges().GetRange() {
				for i := rang.GetBegin(); i <= rang.GetEnd(); i++ {
					ports = append(ports, i)
				}
			}
		}
	}

	return ports[ow.PortUsedSize : len(ports)-1]
}

func (ow *OfferWrapper) CpuRemain() float64 {
	var cpus float64
	for _, res := range ow.Offer.GetResources() {
		if res.GetName() == "cpus" {
			cpus += *res.GetScalar().Value
		}
	}

	return cpus - ow.CpusUsed
}

func (ow *OfferWrapper) MemRemain() float64 {
	var mem float64
	for _, res := range ow.Offer.GetResources() {
		if res.GetName() == "mem" {
			mem += *res.GetScalar().Value
		}
	}

	return mem - ow.MemUsed
}

func (ow *OfferWrapper) DiskRemain() float64 {
	var disk float64
	for _, res := range ow.Offer.GetResources() {
		if res.GetName() == "disk" {
			disk += *res.GetScalar().Value
		}
	}

	return disk - ow.DiskUsed
}
