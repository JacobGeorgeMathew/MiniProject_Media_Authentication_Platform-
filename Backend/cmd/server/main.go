package main

import (
	"fmt"
	"log"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/config"
	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/database"
)

func main() {
	fmt.Println("Hi")

	cfg := config.LoadConfig()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
}
