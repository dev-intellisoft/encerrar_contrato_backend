package routes

import (
	"ec.com/controllers"
	"github.com/gofiber/fiber/v2"
)

func RegistrationRoutes(app *fiber.App) {
	app.Post("/registration", controllers.CustomerCreateSolicitation)
}
