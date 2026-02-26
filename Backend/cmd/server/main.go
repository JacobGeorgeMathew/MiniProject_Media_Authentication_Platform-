package main

import (
	"context"
	"fmt"
	"log"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/config"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/database"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/api/handlers"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/repository"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/services"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/internals/watermark/fingerprint"

	//"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {

	fmt.Println("Hi")

	cfg := config.LoadConfig()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	imageRepo := repository.NewDB(db)

	// Connect to Qdrant
	imageVectorDB, err := fingerprint.NewQdrantDB("localhost", 6334)
	if err != nil {
		log.Fatal("Failed to connect to Qdrant:", err)
	}

	err = imageVectorDB.CreateCollection(context.Background())
	if err != nil {
		log.Println("Collection may already exist:", err)
	}

	imageServices := services.NewImageService(imageRepo, imageVectorDB)

	imageHandler := handlers.NewImageHandler(imageServices)

	fmt.Println("Server initialized successfully")

	app := fiber.New(fiber.Config{
	BodyLimit: 20 * 1024 * 1024, // 20 MB
})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:5173", // Your frontend URLs
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true, // IMPORTANT: Allow cookies to be sent
	}))

	api := app.Group("/api/v1")

	api.Post("/watermark", imageHandler.ImageWatermarkHandler)
	api.Post("/authenticate", imageHandler.ImageAuthHandler)
	// user.Post("/register",func(c *fiber.Ctx) error {

	// 	return c.status(200).JSON(fiber.Map{
	// 		"success": "Under Construction",
	// 	})

	// } )
	// user.Post("/login", func(c *fiber.Ctx) error {

	// 	return c.status(200).JSON(fiber.Map{
	// 		"success": "Under Construction",
	// 	})

	// })
	// user.Post("/logout", func(c *fiber.Ctx) error {

	// 	return c.status(200).JSON(fiber.Map{
	// 		"success": "Under Construction",
	// 	})

	// })

	//user.POST("watermark",)

	fmt.Println("Server Started")
	log.Fatal(app.Listen("localhost:5000"))
}
