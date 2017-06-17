package upstream

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Balancer interface {
	Next([]*Backend) *Backend
}

type rrBalancer struct {
	current int
}

func (b *rrBalancer) Next(bs []*Backend) *Backend {
	if len(bs) == 0 {
		return nil
	}

	if b.current >= len(bs) {
		b.current = 0
	}

	t := bs[b.current]
	b.current++
	return t
}

type weightBalancer struct{}

func (b *weightBalancer) Next(bs []*Backend) *Backend {
	if len(bs) == 0 {
		return nil
	}

	ranges := []float64{0}
	sum := float64(0)
	for _, t := range bs {
		ranges = append(ranges, sum+t.Weight*100)
		sum += t.Weight * 100
	}

	rValue := rand.Float64() * sum
	for i, step := range ranges {
		if step > rValue {
			return bs[i-1]
		}
	}

	return nil
}
