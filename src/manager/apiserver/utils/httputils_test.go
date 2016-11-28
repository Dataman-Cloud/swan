package utils

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestCheckForJSON(t *testing.T) {
	r1, _ := http.NewRequest("POST", "/", nil)
	header := map[string][]string{
		"Content-Type": {"application/json"},
	}

	r1.Header = header
	err1 := CheckForJSON(r1)
	assert.Nil(t, err1)

	r2, _ := http.NewRequest("POST", "/", nil)
	err2 := CheckForJSON(r2)
	assert.NotNil(t, err2)
}
