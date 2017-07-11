package types

const (
	ScaleFailureStop     = "stop"
	ScaleFailureContinue = "continue"
)

type ScalePolicy struct {
	Instances int
	IPs       []string
	Step      int
	OnFailure string
}
