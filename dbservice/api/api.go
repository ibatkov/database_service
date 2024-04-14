package api

import (
	"context"
	"database-service/dbservice/api/config"
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"net/http"
	"time"
)

type Api struct {
	config *config.Values
	router *gin.Engine
	db     *sql.DB
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
	api.config, err = config.ReadConfig()
	if err != nil {
		return
	}

	api.db, err = NewDB(api.config)
	if err != nil {
		return
	}

	api.Init()

	return api, nil

}

func NewDB(config *config.Values) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.GetDSN())
	if err != nil {
		return nil, err
	}
	timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	err = db.PingContext(timeout)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
