package apiserver

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/src/apiserver/metrics"
	"github.com/Dataman-Cloud/swan/src/config"

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
	addr         string
	leaderAddr   string
	apiRegisters []ApiRegister
}

func init() {
	metrics.Register()
}

func NewApiServer(addr string) *ApiServer {
	return &ApiServer{
		addr: addr,
	}
}

func Install(apiServer *ApiServer, apiRegister ApiRegister) {
	apiServer.apiRegisters = append(apiServer.apiRegisters, apiRegister)
}

func (apiServer *ApiServer) UpdateLeaderAddr(addr string) {
	apiServer.leaderAddr = addr
}

func (apiServer *ApiServer) Start() error {
	wsContainer := restful.NewContainer()

	// Register webservices here
	for _, ws := range apiServer.apiRegisters {
		ws.Register(wsContainer)
	}

	// Add container filter to enable CORS
	cors := restful.CrossOriginResourceSharing{
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH"},
		CookiesAllowed: false,
		Container:      wsContainer}
	wsContainer.Filter(cors.Filter)

	// Add log filter
	wsContainer.Filter(NCSACommonLogFormatLogger())

	// proxy the request to leader server
	wsContainer.Filter(apiServer.proxy())

	// Add prometheus metrics
	wsContainer.Handle("/metrics", promhttp.Handler())

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs and enter http://localhost:8080/apidocs.json in the api input field.
	// TODO(xychu): add a config flag for swagger UI, and also for the swagger UI file path.
	swggerUiPath, _ := filepath.Abs("./third_party/swagger-ui-2.2.8")
	logrus.Debugf("xychu:  swaggerUIPath: %s", swggerUiPath)
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

	logrus.Printf("start listening on %s", apiServer.addr)
	server := &http.Server{Addr: apiServer.addr, Handler: wsContainer}
	logrus.Fatal(server.ListenAndServe())

	return nil
}

func NCSACommonLogFormatLogger() restful.FilterFunction {
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

func (apiServer *ApiServer) proxy() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		if apiServer.leaderAddr == apiServer.addr {
			chain.ProcessFilter(req, resp)
			return
		}

		r := req.Request

		leaderUrl, err := url.Parse(apiServer.leaderAddr + r.URL.Path)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}

		if leaderUrl.Scheme == "" {
			leaderUrl.Scheme = "http"
		}

		rr, err := http.NewRequest(r.Method, leaderUrl.String(), r.Body)
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
