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

	// ── Config ────────────────────────────────────────────────────────────────
	cfg := config.LoadConfig()

	// ── Database (pgxpool — required by pgvector) ─────────────────────────────
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// ── Repositories ──────────────────────────────────────────────────────────
	imageRepo := repository.NewImageRepo(db)
	//userRepo := repository.NewUserRepo(db)

	// ── Services ──────────────────────────────────────────────────────────────
	imageService := services.NewImageService(imageRepo)
	userService := services.NewUserService(imageRepo, 0) // 0 → bcrypt.DefaultCost

	// ── Handlers ─────────────────────────────────────────────────────────────
	imageHandler := handlers.NewImageHandler(imageService)
	userHandler := handlers.NewUserHandler(userService)

	// ── Fiber app ────────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024 * 1024, // 20 MB
	})

	// ── Global middleware ─────────────────────────────────────────────────────
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} ${method} ${path} — ${latency}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// ── Routes ───────────────────────────────────────────────────────────────
	api := app.Group("/api/v1")

	// ── Health check ──────────────────────────────────────────────────────────
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	})

	// ── Public: image verification (no login required) ────────────────────────
	// Anyone can POST an image here to look up its ownership / provenance.
	api.Post("/verify", imageHandler.ImageVerifyHandler)

	// ── Public: user auth ─────────────────────────────────────────────────────
	users := api.Group("/users")
	users.Post("/register", userHandler.Register)
	users.Post("/login", userHandler.Login)

	// ── Protected: user profile management ───────────────────────────────────
	me := users.Group("/me", middleware.AuthMiddleware)
	me.Get("/", userHandler.GetMe)
	me.Put("/", userHandler.UpdateProfile)
	me.Put("/password", userHandler.ChangePassword)
	me.Delete("/", userHandler.DeactivateAccount)

	// Admin / lookup (add a role-check middleware in production).
	users.Get("/:id", userHandler.GetUserByID)

	// ── Protected: image operations ───────────────────────────────────────────
	images := api.Group("/images", middleware.AuthMiddleware)
	images.Post("/watermark", imageHandler.ImageWatermarkHandler)
	images.Post("/authenticate", imageHandler.ImageAuthHandler)

	// ── Start ─────────────────────────────────────────────────────────────────
	fmt.Println("Server started on http://localhost:5000")
	log.Fatal(app.Listen("localhost:5000"))
}
