package janitor

import (
	"math/rand"
	"time"
)

type WeightLoadBalancer struct {
}

func NewWeightLoadBalancer() *WeightLoadBalancer {
	return &WeightLoadBalancer{}
}

func (rr *WeightLoadBalancer) Seed(targets []*Target) *Target {
	// protect targets from being empty
	if len(targets) == 0 {
		return nil
	}

	rand.Seed(int64(time.Now().Nanosecond()))
	ranges := []float64{0}
	previousSum := float64(0)
	for _, t := range targets {
		ranges = append(ranges, previousSum+t.Weight*100)
		previousSum += t.Weight * 100
	}

	rValue := rand.Float64() * previousSum
	for i, step := range ranges {
		if step > rValue {
			return targets[i-1]
		}
	}

	return nil
}
