package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/src/manager/framework/engine"
	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

type Api struct {
	port   string
	Engine *engine.Engine
}

func NewApi(eng *engine.Engine) *Api {
	return &Api{
		port:   ":12306",
		Engine: eng,
	}
}

func (api *Api) Start(ctx context.Context) error {
	router := gin.Default()

	router.GET("/v-try/apps", api.ListApp)
	router.POST("/v-try/apps", api.CreateApp)
	router.GET("/v-try/apps/:app_id", api.GetApp)
	router.Run(api.port)
	return nil
}

func (api *Api) CreateApp(c *gin.Context) {
	var version types.Version

	if c.BindJSON(&version) == nil && CheckVersion(&version) == nil {
		err := api.Engine.CreateApp(&version)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "version accepted"})
		}
	} else {
		c.JSON(400, gin.H{"status": "unauthorized"})
	}
}

func (api *Api) ListApp(c *gin.Context) {
}

func (api *Api) GetApp(c *gin.Context) {
	app, err := api.Engine.InspectApp(c.Param("app_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
	} else {
		c.JSON(http.StatusOK, gin.H{"app": app.AsJson()})
	}
}

func CheckVersion(version *types.Version) error {
	// image format
	// mode valid
	// instance exists
	return nil
}
