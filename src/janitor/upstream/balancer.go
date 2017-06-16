package upstream

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Balancer interface {
	Next([]*Target) *Target
}

type rrBalancer struct {
	current int
}

func (b *rrBalancer) Next(ts []*Target) *Target {
	if len(ts) == 0 {
		return nil
	}

	if b.current >= len(ts) {
		b.current = 0
	}

	t := ts[b.current]
	b.current++
	return t
}

type weightBalancer struct{}

func (b *weightBalancer) Next(ts []*Target) *Target {
	if len(ts) == 0 {
		return nil
	}

	ranges := []float64{0}
	sum := float64(0)
	for _, t := range ts {
		ranges = append(ranges, sum+t.Weight*100)
		sum += t.Weight * 100
	}

	rValue := rand.Float64() * sum
	for i, step := range ranges {
		if step > rValue {
			return ts[i-1]
		}
	}

	return nil
}
