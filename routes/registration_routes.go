package routes

import (
	"ec.com/controllers"
	"github.com/gofiber/fiber/v2"
)

func RegistrationRoutes(app *fiber.App) {
	app.Get("/registration/services", controllers.GetRegistrationServices)
	app.Get("/registration/agencies/:id", controllers.GetAgencyById)
	app.Get("/agency/logo/:id", controllers.GetAgencyById)
	app.Post("/registration", controllers.CustomerCreateSolicitation)
	app.Post("/site/checkout", controllers.CreateSiteCheckout)
}
