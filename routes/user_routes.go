package routes

import (
	"ec.com/controllers"
	"github.com/gofiber/fiber/v2"
)

func UserRoutes(app *fiber.App) {
	app.Get("/users/", controllers.GetUsers)
	app.Get("users/me", controllers.Me)
	app.Post("/users/", controllers.CreateUser)
}
