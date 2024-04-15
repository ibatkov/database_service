package api

import (
	"context"
	"database-service/dbservice/api/config"
	"database-service/dbservice/api/controller"
	"database-service/dbservice/api/logger"
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"time"
)

type Api struct {
	config *config.Values
	router *gin.Engine
	db     *sql.DB
	logger logger.Logger
}

func (api *Api) Run() {
	err := api.router.Run()
	if err != nil {
		panic(err)
	}
}

func (api *Api) InitRoutes() {
	api.router = gin.New()

	dataController := controller.BuildDataController(api.logger, api.db, api.config)
	dataController.AddRoutes(api.router)
}

func NewApi() (api Api, err error) {
	api.config, err = config.ReadConfig()
	if err != nil {
		return
	}

	l, err := zap.NewProduction()
	if err != nil {
		return
	}
	api.logger = l.Sugar()

	api.db, err = NewDB(api.config)
	if err != nil {
		api.logger.Error(err)
		return
	}

	api.InitRoutes()

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
