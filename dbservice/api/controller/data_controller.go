package controller

import (
	"database-service/auth"
	"database-service/dbservice/api/logger"
	"database-service/dbservice/api/response"
	"database-service/dbservice/api/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type DataController struct {
	logger      logger.Logger
	dataService services.Data
	authService auth.Service
}

func (controller *DataController) AddRoutes(router gin.IRoutes) {
	router.GET("/data", controller.GetDataHandler)
}

func NewDataController(authService auth.Service, dataService services.Data, logger logger.Logger) *DataController {
	return &DataController{authService: authService, dataService: dataService, logger: logger}
}

type ResponseBody struct {
	Data []Data `json:"data"`
}

type Data struct {
	Id     int    `json:"id"`
	UserId int    `json:"user_id"`
	Data   string `json:"data"`
}

func (controller *DataController) GetDataHandler(ctx *gin.Context) {
	resp := controller.GetData(ctx.GetHeader("Authorization"))
	if resp.Body == nil {
		ctx.Status(resp.Status)
	}
	ctx.JSON(resp.Status, resp.Body)
}

func (controller *DataController) GetData(bearerToken string) response.Response {
	claims, err := controller.authService.GetClaims(bearerToken)
	if err != nil {
		controller.logger.Error(err)
		return response.Response{
			Status: http.StatusUnauthorized,
			Body:   gin.H{"error": err.Error()},
		}
	}

	data, err := controller.dataService.GetDataByAccessLevel(claims.UserID)
	if err != nil {
		controller.logger.Error(err)
		return response.Response{
			Status: http.StatusInternalServerError,
			Body:   nil,
		}
	}

	var body ResponseBody

	body.Data = make([]Data, 0)
	for _, item := range data {
		body.Data = append(body.Data, Data{
			Id:     item.GetId(),
			UserId: item.GetUserId(),
			Data:   item.GetData(),
		})
	}

	return response.Response{
		Status: http.StatusOK,
		Body:   body,
	}
}
