package new_apiserver

import (
	"net/http"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
)

const (
	API_VERSION = "v_beta"
)

type ApiRegister interface {
	Register(*restful.Container)
}

type ApiServer struct {
	addr         string
	apiRegisters []ApiRegister
}

func NewApiServer(addr string) *ApiServer {
	return &ApiServer{
		addr: addr,
	}
}

func Install(apiServer *ApiServer, apiRegister ApiRegister) {
	apiServer.apiRegisters = append(apiServer.apiRegisters, apiRegister)
}

func (apiServer *ApiServer) Start() error {
	wsContainer := restful.NewContainer()

	// Register webservices here
	for _, ws := range apiServer.apiRegisters {
		ws.Register(wsContainer)
	}

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs and enter http://localhost:8080/apidocs.json in the api input field.
	// TODO(xychu): add a config flag for swagger UI, and also for the swagger UI file path.
	swggerUiPath, _ := filepath.Abs("./swagger-ui-2.2.8")
	logrus.Debugf("xychu:  swaggerUIPath: %s", swggerUiPath)
	config := swagger.Config{
		WebServices: wsContainer.RegisteredWebServices(), // you control what services are visible
		// WebServicesUrl: "",
		ApiVersion: API_VERSION,
		ApiPath:    "/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath:     "/apidocs/",
		SwaggerFilePath: swggerUiPath,
	}
	swagger.RegisterSwaggerService(config, wsContainer)

	logrus.Printf("start listening on %s", apiServer.addr)
	server := &http.Server{Addr: apiServer.addr, Handler: wsContainer}
	logrus.Fatal(server.ListenAndServe())

	return nil
}
