package main

import (
	"github.com/Ganesh-S-Pai/pai-finance/controllers"
	"github.com/Ganesh-S-Pai/pai-finance/initializers"
	"github.com/Ganesh-S-Pai/pai-finance/routes"

	"github.com/iris-contrib/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
)

var app *iris.Application

func init() {
	_ = godotenv.Load()

	initializers.ConnectDB()

	app = iris.New()

	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "https://pai-finance.web.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	app.UseRouter(crs)
}

func main() {
	defer initializers.CloseDB()

	db := initializers.Client.Database("mongo-golang")
	userColl := db.Collection("users")

	router := app.Party("/api/v1")

	routes.AuthRoutes(router)

	adminRoutes := router.Party("/admin", controllers.AuthMiddleware(userColl))
	routes.UserRoutes(adminRoutes)
	vhiwRoutes := router.Party("/vhiw", controllers.AuthMiddleware(userColl))
	routes.SalesLogRoutes(vhiwRoutes)

	app.Listen(":8080")
}
