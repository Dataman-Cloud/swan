package mesos

import (
	"encoding/json"
	"sync"

	mesos "github.com/Dataman-Cloud/swan/mesos/offer"
	"github.com/Dataman-Cloud/swan/mesosproto"
)

type Agent struct {
	id       string
	hostname string
	attrs    []*mesosproto.Attribute

	sync.RWMutex
	offers map[string]*mesos.Offer
}

func newAgent(id, hostname string, attrs []*mesosproto.Attribute) *Agent {
	return &Agent{
		id:       id,
		hostname: hostname,
		attrs:    attrs,
		offers:   make(map[string]*mesos.Offer),
	}
}

func (s *Agent) MarshalJSON() ([]byte, error) {
	s.RLock()
	defer s.RUnlock()

	m := map[string]interface{}{
		"id":         s.id,
		"hostname":   s.hostname,
		"attributes": s.attrs,
		"offers":     s.offers,
	}
	return json.Marshal(m)
}

func (s *Agent) addOffer(offer *mesos.Offer) {
	s.Lock()
	s.offers[offer.GetId()] = offer
	s.Unlock()
}

func (s *Agent) removeOffer(offerID string) bool {
	s.Lock()
	defer s.Unlock()

	found := false
	if _, found = s.offers[offerID]; found {
		delete(s.offers, offerID)
	}
	return found
}

func (s *Agent) empty() bool {
	s.RLock()
	defer s.RUnlock()

	return len(s.offers) == 0
}

func (s *Agent) getOffer(offerId string) *mesos.Offer {
	s.RLock()
	defer s.RUnlock()

	offer, ok := s.offers[offerId]
	if !ok {
		return nil
	}

	return offer
}

func (s *Agent) getOffers() map[string]*mesos.Offer {
	s.RLock()
	defer s.RUnlock()

	return s.offers
}

func (s *Agent) Resources() (cpus, mem, disk float64, ports []uint64) {
	for _, offer := range s.getOffers() {
		cpus += offer.GetCpus()
		mem += offer.GetMem()
		disk += offer.GetDisk()
		ports = append(ports, offer.GetPorts()...)
	}

	return
}

func (s *Agent) Attributes() map[string]string {
	attrs := make(map[string]string)

	for _, offer := range s.getOffers() {
		for k, v := range offer.GetAttrs() {
			attrs[k] = v
		}
	}

	return attrs
}
