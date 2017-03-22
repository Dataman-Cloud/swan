package connector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConnector(t *testing.T) {
	c := NewConnector()
	assert.NotNil(t, c)
}

func TestInstance(t *testing.T) {
	c := Instance()
	assert.Equal(t, c, NewConnector())
}

func TestInstanceTwice(t *testing.T) {
	assert.Equal(t, Instance(), Instance())
}
