package types

type UpdateBody struct {
	Instances int     `json:"instances"`
	Canary    *canary `json:"canary"`
}

type canary struct {
	Enabled bool    `json:"enabled"`
	Value   float64 `json:"value"`
}

type UpdateWeightBody struct {
	Weight float64 `json:"weight"`
}

type UpdateWeightsBody struct {
	Weights map[string]float64 `json:"weights"`
}
