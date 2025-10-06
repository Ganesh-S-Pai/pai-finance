package routes

import (
	"github.com/Ganesh-S-Pai/pai-finance/controllers"
	"github.com/Ganesh-S-Pai/pai-finance/initializers"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
)

func UserRoutes(rg router.Party) {
	router := rg.Party("/users")

	db := initializers.Client.Database("mongo-golang")
	userColl := db.Collection("users")

	userC := controllers.NewUserController(userColl)

	router.Use(iris.Compression)
	router.Get("/", userC.GetAllUsers)
	router.Get("/{id}", userC.GetUserByID)
	router.Post("/", userC.AddUser)
	router.Put("/{id}", userC.UpdateUserByID)
	router.Delete("/{id}", userC.DeleteUserByID)
}
