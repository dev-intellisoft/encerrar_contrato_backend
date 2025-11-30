package routes

import (
	"ec.com/controllers"
	"github.com/gofiber/fiber/v2"
)

func ServiceRoutes(app *fiber.App) {
	app.Get("/services", controllers.GetServices)
	app.Get("/services/:id", controllers.GetServiceById)
	app.Post("/services", controllers.CreateService)
	app.Put("/services/:id", controllers.UpdateService)
	app.Delete("/services/:id", controllers.DeleteService)
}
