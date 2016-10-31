package health

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTCPChecker(t *testing.T) {
	checker := NewTCPChecker("xxxxx", "x.x.x.x:yyyy", 8080, 3, 5, 5, nil, "xxx", "yyy")
	assert.Equal(t, checker.ID, "xxxxx")
}

func TestTCPCheckerStart(t *testing.T) {
	checker := NewTCPChecker("xxxxx", "x.x.x.x:yyyy", 8080, 3, 5, 5, nil, "xxx", "yyy")
	checker.Start()
}

func TestTCPCheckerStop(t *testing.T) {
	checker := NewTCPChecker("xxxxx", "x.x.x.x:yyyy", 8080, 3, 5, 5, nil, "xxx", "yyy")
	checker.Stop()
}
