package types

const (
	DefaultCanaryUpdateDelay = 5

	CanaryUpdateOnFailureStop     = "stop"
	CanaryUpdateOnFailureContinue = "continue"
)

type CanaryUpdateBody struct {
	Version   *Version `json:"version"`
	Instances int      `json:"instances"`
	Value     float64  `json:"value"`
	OnFailure string   `json:"onFailure"`
	Delay     float64  `json:"delay"`
}
