package api

import "github.com/gofiber/fiber/v2"

func SetupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Log Viewer",
		})
	})

	app.Get("/api/logs/:id", GetLogByID)
	app.Post("/api/logs/register", RegisterLogFile)
	app.Get("/api/all/logs", GetRegisteredFiles)
	app.Delete("/api/logs/files/:id", DeleteLogFile)
}
