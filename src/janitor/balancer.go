package janitor

import "math/rand"

type Balancer interface {
	Next([]*Target) *Target
}

type RoundRobinBalancer struct {
	current int
}

func (b *RoundRobinBalancer) Next(ts []*Target) *Target {
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

type WeightBalancer struct{}

func (b *WeightBalancer) Next(ts []*Target) *Target {
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
