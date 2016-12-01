package api

import (
	"net/http"

	"github.com/Dataman-Cloud/swan/src/types"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

type Api struct {
	port string
}

func NewApi() *Api {
	return &Api{
		port: ":12306",
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
		c.JSON(http.StatusOK, gin.H{"status": "version accepted"})
	} else {
		c.JSON(400, gin.H{"status": "unauthorized"})
	}
}

func (api *Api) ListApp(c *gin.Context) {
}

func (api *Api) GetApp(c *gin.Context) {
}

func CheckVersion(version *types.Version) error {
	// image format
	// mode valid
	// instance exists
	return nil
}
