package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/src/config"
	"github.com/Dataman-Cloud/swan/src/manager/api/server/metrics"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// RootPaths lists the paths available at root.
// For example: "/healthz", "/apis".
type RootPaths struct {
	// paths are the paths available at root.
	Paths []string
}

type ApiRegister interface {
	Register(*restful.Container)
}

type ApiServer struct {
	listenAddr string
	// TODO(nmg): use `leaderAddr` to update leader addr for request proxy.
	// request should be proxy from follower to leader.
	leaderAddr string

	registers []ApiRegister
	server    *http.Server
}

func init() {
	metrics.Register()
}

func NewApiServer(l string) *ApiServer {
	return &ApiServer{
		listenAddr: l,
	}
}

func Install(s *ApiServer, r ApiRegister) {
	s.registers = append(s.registers, r)
}

func (s *ApiServer) UpdateLeaderAddr(addr string) {
	s.leaderAddr = addr
}

func (s *ApiServer) Start() error {
	wsContainer := restful.NewContainer()

	// Register webservices here
	for _, r := range s.registers {
		r.Register(wsContainer)
	}

	// Add container filter to enable CORS
	cors := restful.CrossOriginResourceSharing{
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH"},
		CookiesAllowed: false,
		Container:      wsContainer}
	wsContainer.Filter(cors.Filter)

	// Add log filter
	wsContainer.Filter(s.NCSACommonLogFormatLogger())

	//
	wsContainer.Filter(s.ProxyRequest())

	// Add prometheus metrics
	wsContainer.Handle("/metrics", promhttp.Handler())

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs and enter http://localhost:8080/apidocs.json in the api input field.
	// TODO(xychu): add a config flag for swagger UI, and also for the swagger UI file path.
	swggerUiPath, _ := filepath.Abs("./3rdparty/swagger-ui-2.2.8")
	config := swagger.Config{
		WebServices: wsContainer.RegisteredWebServices(), // you control what services are visible
		// WebServicesUrl: "",
		ApiVersion: config.API_PREFIX,
		ApiPath:    "/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath:     "/apidocs/",
		SwaggerFilePath: swggerUiPath,
	}
	swagger.RegisterSwaggerService(config, wsContainer)

	// Add API index handler
	wsContainer.ServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		if r.URL.Path != "/" && r.URL.Path != "/index.html" {
			// Since "/" matches all paths, handleIndex is called for all paths for which there is no handler registered.
			// We want to return a 404 status with a list of all valid paths, incase of an invalid URL request.
			status = http.StatusNotFound
			w.WriteHeader(status)
			return
		}
		var handledPaths []string
		// Extract the paths handled using restful.WebService
		for _, ws := range wsContainer.RegisteredWebServices() {
			handledPaths = append(handledPaths, ws.RootPath())
		}

		// Extract the paths handled using mux handler.
		handledPaths = append(handledPaths, "/metrics", "/apidocs/")
		sort.Strings(handledPaths)

		output, err := json.MarshalIndent(RootPaths{Paths: handledPaths}, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(output)
	})

	logrus.Printf("apiserver listening on %s", s.listenAddr)

	srv := &http.Server{Addr: s.listenAddr, Handler: wsContainer}
	s.server = srv

	return srv.ListenAndServe()
}

// gracefully shutdown.
func (s *ApiServer) Shutdown() error {
	// If s.server is nil, api server is not running.
	if s.server != nil {
		// NOTE(nmg): need golang 1.8+ to run this method.
		return s.server.Shutdown(nil)
	}

	return nil
}

func (s *ApiServer) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}

	return nil
}

func (s *ApiServer) NCSACommonLogFormatLogger() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		var username = "-"
		if req.Request.URL.User != nil {
			if name := req.Request.URL.User.Username(); name != "" {
				username = name
			}
		}
		chain.ProcessFilter(req, resp)
		logrus.Printf("%s - %s [%s] \"%s %s %s\" %d %d",
			strings.Split(req.Request.RemoteAddr, ":")[0],
			username,
			time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			req.Request.Method,
			req.Request.URL.RequestURI(),
			req.Request.Proto,
			resp.StatusCode(),
			resp.ContentLength(),
		)
	}
}

func (s *ApiServer) ProxyRequest() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		if s.leaderAddr == s.listenAddr {
			chain.ProcessFilter(req, resp)
			return
		}

		r := req.Request

		// NOTE(nmg): If you just use ip address here, the `url.Parse` with get error with
		// `first path segment in URL cannot contain colon`.
		// It's golang 1.8's bug. more details see https://github.com/golang/go/issues/18824.
		leaderUrl := s.leaderAddr
		if !strings.HasPrefix(s.leaderAddr, "http://") {
			leaderUrl = "http://" + s.leaderAddr
		}

		leaderURL, err := url.Parse(leaderUrl + r.URL.Path)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}

		rr, err := http.NewRequest(r.Method, leaderURL.String(), r.Body)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}

		copyHeader(r.Header, &rr.Header)

		// Create a client and query the target
		var transport http.Transport
		leaderResp, err := transport.RoundTrip(rr)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}

		defer leaderResp.Body.Close()
		body, err := ioutil.ReadAll(leaderResp.Body)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}

		dH := resp.Header()
		copyHeader(leaderResp.Header, &dH)
		dH.Add("Requested-Host", rr.Host)

		resp.Write(body)
		return
	}
}

func copyHeader(source http.Header, dest *http.Header) {
	for n, v := range source {
		for _, vv := range v {
			dest.Add(n, vv)
		}
	}
}
