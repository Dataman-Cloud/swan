package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsManager(t *testing.T) {
	n := Node{Role: RoleManager}

	assert.True(t, n.IsManager())
}

func TestIsManagerFalse(t *testing.T) {
	n := Node{Role: RoleAgent}

	assert.False(t, n.IsManager())
}

func TestIsAgent(t *testing.T) {
	n := Node{Role: RoleAgent}

	assert.True(t, n.IsAgent())
}

func TestIsAgentFalse(t *testing.T) {
	n := Node{Role: RoleManager}

	assert.False(t, n.IsAgent())
}
