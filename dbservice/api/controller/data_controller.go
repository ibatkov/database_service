package controller

import (
	"database-service/auth"
	"database-service/dbservice/api/logger"
	"database-service/dbservice/api/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type DataController struct {
	logger      logger.Logger
	service     services.Data
	authService auth.Service
}

func (controller *DataController) AddRoutes(router gin.IRoutes) {
	router.GET("/data", controller.getData)
}

func NewDataController(authService auth.Service, service services.Data, logger logger.Logger) *DataController {
	return &DataController{authService: authService, service: service, logger: logger}
}

type Response struct {
	Data []Data `json:"data"`
}

type Data struct {
	Id     int    `json:"id"`
	UserId int    `json:"user_id"`
	Data   string `json:"data"`
}

func (controller *DataController) getData(ctx *gin.Context) {
	claims, err := controller.authService.GetClaims(ctx.GetHeader("Authorization"))
	if err != nil {
		controller.logger.Error(err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	data, err := controller.service.GetDataByAccessLevel(claims.UserID)
	if err != nil {
		controller.logger.Error(err)
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	var response Response

	response.Data = make([]Data, 0)
	for _, item := range data {
		response.Data = append(response.Data, Data{
			Id:     item.GetId(),
			UserId: item.GetUserId(),
			Data:   item.GetData(),
		})
	}

	ctx.JSON(http.StatusOK, response)
}
