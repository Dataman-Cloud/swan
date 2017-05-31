package types

import (
	"fmt"
)

var supportedOperator = []string{"==", "!=", "~="}

type Constraint struct {
	Attribute string `json:"attribute"`
	Operator  string `json:"operator"`
	Value     string `json:"value"`
}

func (c *Constraint) validate() error {
	for _, str := range supportedOperator {
		if str == c.Operator {
			return nil
		}
	}

	return fmt.Errorf("Operator not supported. supported operators is %v", supportedOperator)
}

func (c *Constraint) Match(attrs map[string]string) bool {
	for k, v := range attrs {
		if k == c.Attribute {
			switch c.Operator {
			case "==":
				return equal(c.Value, v)
			case "!=":
				return not(c.Value, v)
			case "~=":
				return like(c.Value, v)
			}
		}
	}

	return false
}

func equal(n, m string) bool {
	return n == m
}

func not(n, m string) bool {
	return n != m
}

func like(n, m string) bool {
	return true
}
