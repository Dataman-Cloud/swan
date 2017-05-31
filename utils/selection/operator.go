package selection

/*
	this pkgs is from https://github.com/kubernetes/kubernetes.
	thanks for the great project
*/

// Operator represents a key/field's relationship to value(s).
// See labels.Requirement and fields.Requirement for more details.
type Operator string

const (
	DoesNotExist Operator = "!"
	Equals       Operator = "="
	DoubleEquals Operator = "=="
	In           Operator = "in"
	NotEquals    Operator = "!="
	NotIn        Operator = "notin"
	Exists       Operator = "exists"
	GreaterThan  Operator = "gt"
	LessThan     Operator = "lt"
)
