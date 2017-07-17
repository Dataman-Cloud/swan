package types

type UpdateWeightBody struct {
	Weight float64 `json:"weight"`
}

type UpdateWeightsBody struct {
	Weights map[string]float64 `json:"weights"`
}
