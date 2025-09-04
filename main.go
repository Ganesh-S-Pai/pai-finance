package main

import (
	"vhiw-sales-log/initializers"
	"vhiw-sales-log/routes"

	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
)

var app *iris.Application

func init() {
	_ = godotenv.Load()

	initializers.ConnectDB()

	app = iris.New()
}

func main() {
	defer initializers.CloseDB()

	router := app.Party("/api/v1/vhiw")

	routes.SalesLogRoutes(router)

	app.Listen(":8080")
}
