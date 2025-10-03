package main

import (
	"vhiw-sales-log/initializers"
	"vhiw-sales-log/routes"

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

	router := app.Party("/api/v1/vhiw")

	routes.SalesLogRoutes(router)

	app.Listen(":8080")
}
