package routes

import (
	"ec.com/controllers"
	"github.com/gofiber/fiber/v2"
)

func RegistrationRoutes(app *fiber.App) {
	app.Get("/registration/services", controllers.GetRegistrationServices)
	app.Post("/registration", controllers.CustomerCreateSolicitation)
	app.Post("/site/checkout", controllers.CreateSiteCheckout)
}
