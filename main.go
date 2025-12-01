package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"ec.com/auth"
	"ec.com/database"
	m "ec.com/models"
	"ec.com/routes"
	s "ec.com/storage"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed to load configuration file ")
	}
	//pkg.InitNode()
	database.Connect()
}

func jwtError(c *fiber.Ctx, err error) error {
	println(err.Error())
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": "Unauthorized",
	})
}

func main() {
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	tokenStore := s.NewGormTokenStore(database.DB)
	manager.MustTokenStorage(tokenStore, nil)
	srv := server.NewDefaultServer(manager)
	clientStore := store.NewClientStore()
	manager.MapClientStorage(clientStore)
	manager.MapAccessGenerate(
		generates.NewJWTAccessGenerate("auth-server", []byte("SECRET_SIGNING_KEY"), jwt.SigningMethodHS256),
	)

	clientStore.Set("encerrar", &models.Client{
		ID:     "encerrar",
		Secret: "contract123",
		Domain: "http://localhost",
	})

	manager.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)

	srv.SetPasswordAuthorizationHandler(func(ctx context.Context, clientID, username, password string) (string, error) {
		log.Printf("PasswordAuthHandler: client=%s user=%s pass=%s", clientID, username, password)

		var user m.User
		var err error

		if len(password) == 4 {
			log.Println("Trying validation code login")
			user, err = auth.GetUserWithValidaCode(username, password)
		} else {
			log.Println("Trying password login")
			user, err = auth.GetUserWithPassword(username, password)
		}

		if err != nil {
			log.Println("Auth failed:", err)
			return "", errors.ErrAccessDenied
		}

		userJSON, _ := json.Marshal(struct {
			ID     uuid.UUID
			Agency string
			Email  string
		}{
			ID:     user.ID,
			Agency: user.Agency,
			Email:  user.Email,
		})

		log.Println("Auth success:", string(userJSON))
		return string(userJSON), nil
	})

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // or "http://localhost:58499"
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	routes.RegistrationRoutes(app)
	routes.PaymentRoutes(app)

	app.Static("/", "./public")
	app.Static("/angecies", "./uploads/angecies")

	app.Post("/token", func(c *fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			err := srv.HandleTokenRequest(w, r)
			if err != nil {
				println(err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}))(c.Context())

		return nil
	})

	app.Post("/validation/code", func(c *fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := srv.HandleTokenRequest(w, r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}))(c.Context())

		return nil
	})

	app.Use("/", jwtware.New(jwtware.Config{
		SigningKey:   []byte("SECRET_SIGNING_KEY"),
		ContextKey:   "user", // onde o token será armazenado no contexto
		ErrorHandler: jwtError,
	}))

	routes.UserRoutes(app)
	routes.SolicitationRoutes(app)
	routes.AgencyRoutes(app)
	routes.ServiceRoutes(app)

	port := os.Getenv("PORT")

	if err := app.Listen(":" + port); err != nil {
		println(err.Error())
	}
}
