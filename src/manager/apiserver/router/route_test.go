package router

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Handler(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func TestnewRoute(t *testing.T) {
	r := NewRoute("PUT", "/foobar", Handler)
	assert.Equal(t, "PUT", r.Method())
	assert.Equal(t, "/foobar", r.Path())
	assert.Equal(t, Handler, r.Handler())
}
