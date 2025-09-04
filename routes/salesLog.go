package routes

import (
	"vhiw-sales-log/controllers"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
)

func SalesLogRoutes(rg router.Party) {
	router := rg.Party("/sales-logs")

	router.Use(iris.Compression)
	router.Get("/", controllers.GetSalesLogs)
	router.Get("/{id}", controllers.GetSalesLogByID)
	router.Post("/", controllers.AddSalesLog)
	router.Put("/{id}", controllers.UpdateSalesLogByID)
	router.Delete("/{id}", controllers.DeleteSalesLogByID)
}
