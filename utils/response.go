package utils

import (
	"github.com/kataras/iris/v12"
)

type APIResponse struct {
	Status  string      `json:"status"`         
	Message string      `json:"message"`        
	Data    interface{} `json:"data,omitempty"` 
}

func SendResponse(ctx iris.Context, httpCode int, status string, message string, data interface{}) {
	response := APIResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}

	ctx.StatusCode(httpCode)
	ctx.JSON(response)
}

func SendResponseAndStop(ctx iris.Context, httpCode int, status string, message string, data interface{}) {
	ctx.StopWithJSON(httpCode, iris.Map{"status": status, "message": message, "data": data})
}
