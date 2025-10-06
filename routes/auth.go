package routes

import (
	"github.com/Ganesh-S-Pai/pai-finance/controllers"
	"github.com/Ganesh-S-Pai/pai-finance/initializers"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
)

func AuthRoutes(rg router.Party) {
	router := rg.Party("/auth")

	db := initializers.Client.Database("mongo-golang")
	userColl := db.Collection("users")

	authC := controllers.NewAuthController(userColl)

	router.Use(iris.Compression)
	router.Post("/signup", authC.Signup)
	router.Post("/login", authC.Login)
}
