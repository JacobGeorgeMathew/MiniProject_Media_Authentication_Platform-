package main

import (
	"fmt"
	"log"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/config"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/database"

	"github.com/gofiber/fiber"
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

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:5173", // Your frontend URLs
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true, // IMPORTANT: Allow cookies to be sent
	}))

	api := app.Group("/api")

	user := api.Group("/student")
	user.Post("/register",func(c *fiber.Ctx) error {

		return c.status(200).JSON(fiber.Map{
			"success": "Under Construction",
		})
		
	} )
	user.Post("/login", func(c *fiber.Ctx) error {

		return c.status(200).JSON(fiber.Map{
			"success": "Under Construction",
		})
		
	})
	user.Post("/logout", func(c *fiber.Ctx) error {

		return c.status(200).JSON(fiber.Map{
			"success": "Under Construction",
		})
		
	})

	//user.POST("watermark",)

	fmt.Println("Server Started")
	log.Fatal(app.Listen("localhost:5000"))
}
