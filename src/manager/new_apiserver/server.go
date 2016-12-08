package new_apiserver

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	//"github.com/emicklei/go-restful/swagger"
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

	//// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	//// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	//// Open http://localhost:8080/apidocs and enter http://localhost:8080/apidocs.json in the api input field.
	//config := swagger.Config{
	//	WebServices:    wsContainer.RegisteredWebServices(), // you control what services are visible
	//	WebServicesUrl: "http://localhost:8080",
	//	ApiPath:        "/apidocs.json",
	//
	//	// Optionally, specifiy where the UI is located
	//	SwaggerPath:     "/apidocs/",
	//	SwaggerFilePath: "/Users/chuxiangyang/go/src/github.com/Dataman-Cloud/swan/example/dist",
	//	}
	//swagger.RegisterSwaggerService(config, wsContainer)

	logrus.Printf("start listening on %s", apiServer.addr)
	server := &http.Server{Addr: apiServer.addr, Handler: wsContainer}
	logrus.Fatal(server.ListenAndServe())

	return nil
}
