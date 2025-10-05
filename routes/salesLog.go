package routes

import (
	"vhiw-sales-log/controllers"
	"vhiw-sales-log/initializers"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
)

func SalesLogRoutes(rg router.Party) {
	router := rg.Party("/sales-logs")

	db := initializers.Client.Database("mongo-golang")
	salesColl := db.Collection("sales-logs")

	salesC := controllers.NewSalesController(salesColl)

	router.Use(iris.Compression)
	router.Get("/", salesC.GetSalesLogs)
	router.Get("/{id}", salesC.GetSalesLogByID)
	router.Post("/", salesC.AddSalesLog)
	router.Put("/{id}", salesC.UpdateSalesLogByID)
	router.Delete("/{id}", salesC.DeleteSalesLogByID)
}
