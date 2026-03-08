package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/config"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/database"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/api/handlers"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/api/middleware"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/repository"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/services"
)

func main() {

	// ── Config & Database ────────────────────────────────────────────────────
	cfg := config.LoadConfig()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// ── Repositories ─────────────────────────────────────────────────────────
	imageRepo := repository.NewImageRepo(db)
	//userRepo := repository.NewUserRepo(db)
	enterpriseRepo := repository.NewEnterpriseRepo(db)

	// ── Services ──────────────────────────────────────────────────────────────
	imageService := services.NewImageService(imageRepo)
	userService := services.NewUserService(imageRepo, 0)
	enterpriseService := services.NewEnterpriseService(enterpriseRepo, 0)

	// ── Handlers ─────────────────────────────────────────────────────────────
	imageHandler := handlers.NewImageHandler(imageService)
	userHandler := handlers.NewUserHandler(userService)
	enterpriseHandler := handlers.NewEnterpriseHandler(enterpriseService, imageService)

	// ── Fiber app ─────────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024 * 1024, // 20 MB
	})

	// ── Global middleware ──────────────────────────────────────────────────────
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} ${method} ${path} — ${latency}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-API-Key",
		AllowCredentials: true,
	}))

	// ── Base route group ──────────────────────────────────────────────────────
	api := app.Group("/api/v1")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	})

	// Public image verification (no JWT, no API key)
	api.Post("/verify", imageHandler.ImageVerifyHandler)

	// ── User routes — public ──────────────────────────────────────────────────
	users := api.Group("/users")
	users.Post("/register", userHandler.Register)
	users.Post("/login", userHandler.Login)
	users.Get("/:id", userHandler.GetUserByID) // add role-check in production

	// ── User routes — JWT protected ───────────────────────────────────────────
	me := users.Group("/me", middleware.AuthMiddleware)
	me.Get("/", userHandler.GetMe)
	me.Put("/", userHandler.UpdateProfile)
	me.Put("/password", userHandler.ChangePassword)
	me.Delete("/", userHandler.DeactivateAccount)

	// ── Image routes — user JWT protected ─────────────────────────────────────
	images := api.Group("/images", middleware.AuthMiddleware)
	images.Post("/watermark", imageHandler.ImageWatermarkHandler)
	images.Post("/authenticate", imageHandler.ImageAuthHandler)

	// =========================================================================
	// Enterprise routes  —  /api/v1/enterprise/...
	// =========================================================================

	ent := api.Group("/enterprise")

	// Public: registration and login (no auth)
	ent.Post("/register", enterpriseHandler.Register)
	ent.Post("/login", enterpriseHandler.Login)

	// Dashboard / management — enterprise JWT
	entMe := ent.Group("/me", middleware.EnterpriseJWTMiddleware)
	entMe.Get("/", enterpriseHandler.GetMe)
	entMe.Put("/", enterpriseHandler.UpdateProfile)
	entMe.Put("/password", enterpriseHandler.ChangePassword)
	entMe.Delete("/", enterpriseHandler.Deactivate)

	// API key management — enterprise JWT
	entKeys := ent.Group("/keys", middleware.EnterpriseJWTMiddleware)
	entKeys.Post("/", enterpriseHandler.CreateAPIKey)
	entKeys.Get("/", enterpriseHandler.ListAPIKeys)
	entKeys.Delete("/:keyId", enterpriseHandler.RevokeAPIKey)

	// Usage / billing dashboard — enterprise JWT
	ent.Get("/usage", middleware.EnterpriseJWTMiddleware, enterpriseHandler.GetUsage)

	// Watermarking-as-a-Service — X-API-Key (called from enterprise backends)
	entAPI := ent.Group("/api", middleware.EnterpriseAPIKeyMiddleware(enterpriseService))
	entAPI.Post("/watermark", enterpriseHandler.WatermarkImage)
	entAPI.Post("/authenticate", enterpriseHandler.AuthenticateImage)

	// ── Start ──────────────────────────────────────────────────────────────────
	fmt.Println("Server started on http://localhost:5000")
	log.Fatal(app.Listen("localhost:5000"))
}
