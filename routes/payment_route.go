package routes

import (
	"ec.com/controllers"
	"github.com/gofiber/fiber/v2"
)

func PaymentRoutes(app *fiber.App) {
	app.Post("/payments", controllers.ProcessPayment)
	app.Post("/payments/credit-card/:solicitation_id", controllers.ProcessCreditCardPayment)
	app.Post("/payments/pix/:solicitation_id", controllers.ProcessPixPayment)
	app.Post("/payments/confirm", controllers.ConfirmPayment)
}
