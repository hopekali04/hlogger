package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/hopekali04/hlogger/api"
)

func main() {
	app := fiber.New(fiber.Config{
		Views: html.New("./views", ".html"),
	})

	app.Use(logger.New())
	app.Static("/", "./public")

	api.SetupRoutes(app)

	log.Fatal(app.Listen(":3000"))
}
