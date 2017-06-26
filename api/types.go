package api

type scaleBody struct {
	instances int      `json:"instances"`
	ips       []string `json:"ips"` // TODO(nmg): Removed after automatic IPAM.
}

type updateWeightBody struct {
	weight float64 `json:"weight"`
}

type updateWeightsBody struct {
	weights map[string]float64 `json:"weights"`
}

type leader struct {
	Leader string `json:"leader"`
	// Follower []string `json:"follower"` // TODO(nmg)
}
