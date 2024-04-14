package api

import (
	"database-service/dbservice/api/config"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Api struct {
	config *config.Config
	router *gin.Engine
}

func (api *Api) Run() {
	err := api.router.Run()
	if err != nil {
		panic(err)
	}
}

func (api *Api) Init() {
	api.router = gin.New()

	api.router.GET("/data",
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "data"})
		},
	)

}

func NewApi() (api Api, err error) {

	api.Init()

	return api, nil

}
