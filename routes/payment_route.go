package routes

import (
	"ec.com/controllers"
	"github.com/gofiber/fiber/v2"
)

func PaymentRoutes(app *fiber.App) {
	app.Post("/payments", controllers.ProcessPayment)
}
