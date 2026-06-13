package routes

import (
	"ec.com/controllers"
	"github.com/gofiber/fiber/v2"
)

func RegistrationRoutes(app *fiber.App) {
	app.Post("/registration/:agency_id", controllers.CustomerCreateSolicitation)
	app.Post("/transfer/:agency_id", controllers.CreateAnTransferSolicitation)
	app.Get("/registration/services", controllers.GetServices)
	app.Get("/agency/logo/:agency_id", controllers.GetAgencyLogo)
	app.Get("/registration/agencies/:id", controllers.GetAgencyById)
	app.Post("/site/checkout", controllers.CreateSiteCheckout)
	app.Post("/site/leads", controllers.CreateSiteLead)
	app.Get("/site/leads/:id/status", controllers.GetSiteLeadStatus)
	app.Post("/site/webhooks/asaas", controllers.HandleAsaasWebhook)
	app.Post("/site/v2/leads", controllers.CreateSiteLeadV2)
	app.Get("/site/v2/leads/:id/status", controllers.GetSiteLeadStatusV2)
	app.Post("/site/v2/webhooks/asaas", controllers.HandleAsaasWebhookV2)
}
