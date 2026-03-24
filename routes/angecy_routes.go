package routes

import (
	"ec.com/controllers"
	"github.com/gofiber/fiber/v2"
)

func AgencyRoutes(app *fiber.App) {
	app.Get("/agencies", controllers.GetAgencies)
	app.Get("/agencies/:id", controllers.GetAgencyById)
	app.Post("/agencies", controllers.CreateAgency)
	app.Post("/agency/transfer", controllers.CreateAgencyTransferSolicitation)
	app.Put("/agencies/:id", controllers.UpdateAgency)
	app.Delete("/agencies/:id", controllers.DeleteAgency)
}
