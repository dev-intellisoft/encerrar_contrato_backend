package routes

import (
	"ec.com/controllers"
	"github.com/gofiber/fiber/v2"
)

func RegistrationRoutes(app *fiber.App) {
	app.Post("/registration/:agency_id", controllers.CustomerCreateSolicitation)
	app.Post("/transfer/:agency_id", controllers.CreateAnTransferSolicitation)
	app.Get("/registration/services", controllers.GetServices)
}
