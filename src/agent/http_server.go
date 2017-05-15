package agent

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"

	"github.com/prometheus/client_golang/prometheus"
)

// HTTPServer instance
type HTTPServer struct {
	listener string
	agentRef *Agent
	engine   *gin.Engine
}

// Deprecated: Use promhttp.Handler instead.
func prometheusHandler() gin.HandlerFunc {
	h := prometheus.UninstrumentedHandler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// NewHTTPServer is singleton instance func
func NewHTTPServer(listener string, a *Agent) *HTTPServer {
	aas := &HTTPServer{
		listener: listener,
		agentRef: a,
	}
	aas.engine = gin.Default()

	aas.engine.GET(aas.agentRef.Janitor.P.MetricsPath, prometheusHandler())

	aas.engine.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	aas.engine.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	aas.engine.GET("/agents", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"agents": aas.agentRef.SerfServer.SerfNode.Members()})
	})

	aas.engine.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"upstreams": aas.agentRef.Janitor.UpstreamLoader.Upstreams})
	})

	return aas
}

// Start func
func (aas *HTTPServer) Start(ctx context.Context, started chan bool) error {
	errCh := make(chan error)
	go func() {
		errCh <- aas.engine.Run(aas.listener)
	}()

	go func() {
		started <- true
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
