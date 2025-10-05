package routes

import (
	"vhiw-sales-log/controllers"
	"vhiw-sales-log/initializers"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
)

func UserRoutes(rg router.Party) {
	router := rg.Party("/auth")

	db := initializers.Client.Database("mongo-golang")
	userColl := db.Collection("users")

	authC := controllers.NewAuthController(userColl)

	router.Use(iris.Compression)
	router.Post("/signup", authC.Signup)
	router.Post("/login", authC.Login)
}
