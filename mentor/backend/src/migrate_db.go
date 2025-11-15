//go:build ignore

package main

import (
	"log"
	"mentor-backend/database"
)

func main() {
	database.Connect()

	log.Println("Starting migration (without foreign keys)...")

	// Migrate models individually without foreign keys

	log.Println("Migration completed!")
}
