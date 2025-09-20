package routes

import (
	"ec.com/controllers"
	"github.com/gofiber/fiber/v2"
)

func SolicitationRoutes(app *fiber.App) {
	app.Post("/solicitations", controllers.CreateSolicitation)
	app.Get("/solicitations", controllers.GetSolicitation)
	app.Get("/solicitations/:id", controllers.GetSolicitationById)
	app.Put("/solicitations/:id", controllers.UpdateSolicitation)
	app.Put("/solicitations/:id/start", controllers.StartSolicitation)
	app.Put("/solicitations/:id/end", controllers.EndSolicitation)
	app.Delete("/solicitations/:id", controllers.DeleteSolicitation)
}
