package mesos

import (
	"encoding/json"

	"github.com/Dataman-Cloud/swan/mesosproto"
)

type Offer struct {
	id         string
	cpus       float64
	mem        float64
	disk       float64
	ports      []uint64
	portRanges []*portRange
	attrs      map[string]string
	hostname   string
	agentId    string
}

type portRange struct {
	begin uint64
	end   uint64
}

func (r *portRange) MarshalJSON() ([]byte, error) {
	return json.Marshal([]uint64{r.begin, r.end})
}

func newOffer(offer *mesosproto.Offer) *Offer {
	f := &Offer{
		id:       offer.GetId().GetValue(),
		hostname: offer.GetHostname(),
		agentId:  offer.GetAgentId().GetValue(),
	}

	var (
		cpus, mem, disk float64
		ports           []uint64
		portRanges      []*portRange
	)

	for _, resource := range offer.Resources {
		if *resource.Name == "cpus" {
			cpus += *resource.Scalar.Value
		}

		if *resource.Name == "mem" {
			mem += *resource.Scalar.Value
		}

		if *resource.Name == "disk" {
			disk += *resource.Scalar.Value
		}

		if *resource.Name == "ports" {
			for _, r := range resource.GetRanges().GetRange() {
				var (
					b = r.GetBegin()
					e = r.GetEnd()
				)

				for i := b; i <= e; i++ {
					ports = append(ports, i)
				}
				portRanges = append(portRanges, &portRange{b, e})
			}
		}

	}

	f.cpus = cpus
	f.mem = mem
	f.disk = disk
	f.ports = ports

	attrs := make(map[string]string, 0)
	for _, attr := range offer.Attributes {
		if attr.GetType() == mesosproto.Value_TEXT {
			attrs[attr.GetName()] = attr.GetText().GetValue()
		}
	}

	f.attrs = attrs

	return f

}

func (f *Offer) GetId() string {
	return f.id
}

func (f *Offer) GetCpus() float64 {
	return f.cpus
}

func (f *Offer) GetMem() float64 {
	return f.mem
}

func (f *Offer) GetDisk() float64 {
	return f.disk
}

func (f *Offer) GetPorts() (ports []uint64) {
	return f.ports
}

func (f *Offer) GetPortRange() (ranges []string) {
	for _, r := range f.portRanges {
		b, _ := json.Marshal(r)

		ranges = append(ranges, string(b))
	}

	return
}

func (f *Offer) GetAgentId() string {
	return f.agentId
}

func (f *Offer) GetAttrs() map[string]string {
	return f.attrs
}

func (f *Offer) GetHostname() string {
	return f.hostname
}

func (f *Offer) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"id":       f.id,
		"cpus":     f.cpus,
		"mem":      f.mem,
		"disk":     f.disk,
		"ports":    f.portRanges,
		"hostname": f.hostname,
		"attrs":    f.attrs,
	}

	return json.Marshal(m)
}

func (f *Offer) Ports() func() uint64 {
	ch := make(chan uint64, 1)
	go func() {
		for _, port := range f.ports {
			ch <- port
		}

		close(ch)
	}()

	fn := func() uint64 {
		port, ok := <-ch
		if !ok {
			return 0
		}

		return port
	}

	return fn
}
