package offer

import (
	"encoding/json"

	"github.com/Dataman-Cloud/swan/mesosproto"
)

type Offer struct {
	id       string
	cpus     float64
	mem      float64
	disk     float64
	ports    []*portRange
	attrs    map[string]string
	hostname string
	agentId  *mesosproto.AgentID
}

type portRange struct {
	begin uint64
	end   uint64
}

func (r *portRange) MarshalJSON() ([]byte, error) {
	return json.Marshal([]uint64{r.begin, r.end})
}

func NewOffer(offer *mesosproto.Offer) *Offer {
	f := &Offer{
		id:       offer.GetId().GetValue(),
		hostname: offer.GetHostname(),
		agentId:  offer.GetAgentId(),
	}

	var (
		cpus, mem, disk float64
		ports           []*portRange
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
			for _, rang := range resource.GetRanges().GetRange() {
				ports = append(ports, &portRange{rang.GetBegin(), rang.GetEnd()})
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
	for _, r := range f.ports {
		for i := r.begin; i <= r.end; i++ {
			ports = append(ports, i)
		}
	}

	return
}

func (f *Offer) GetPortRange() (ranges []string) {
	for _, r := range f.ports {
		b, _ := json.Marshal(r)

		ranges = append(ranges, string(b))
	}

	return
}

func (f *Offer) GetAgentId() *mesosproto.AgentID {
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
		"ports":    f.ports,
		"hostname": f.hostname,
		"attrs":    f.attrs,
	}

	return json.Marshal(m)
}
