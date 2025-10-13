package routes

import (
	"github.com/Ganesh-S-Pai/pai-finance/controllers"
	"github.com/Ganesh-S-Pai/pai-finance/initializers"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
)

func StatementRoutes(rg router.Party) {
	router := rg.Party("/statement")

	db := initializers.Client.Database("mongo-golang")
	airbnbColl := db.Collection("airbnb")

	airbnbC := controllers.NewAirbnbController(airbnbColl)

	router.Use(iris.Compression)
	router.Get("/", airbnbC.GetAllStatements)
	router.Post("/", airbnbC.AddStatement)
	router.Put("/{id}", airbnbC.UpdateStatementByID)
	router.Delete("/{id}", airbnbC.DeleteStatementByID)
}
