package middleware

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/api"
)

type CORSMiddleware struct{}

func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{}
}

func (c *CORSMiddleware) Name() string {
	return "cors"
}

// WrapHandler returns a new handler function wrapping the previous one in the request chain.
func (c *CORSMiddleware) WrapHandler(handler api.HandlerFunc) api.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, X-Registry-Auth")
		w.Header().Add("Access-Control-Allow-Methods", "HEAD, GET, POST, DELETE, PUT, OPTIONS")
		return handler(w, r)
	}
}
