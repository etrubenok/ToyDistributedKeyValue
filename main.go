package main

import (
	"ToyDistributedKeyValue/handlers"
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {
	// Initialize RocksDB
	handlers.InitDB()
	defer handlers.CloseDB()

	// Initialize a new Fiber app
	app := fiber.New()

	// Route to health check
	app.Get("/healthcheck", handlers.HealthCheck)

	// Route to handle POST /key
	app.Post("/key", handlers.SetKeyValue)

	// Route to handle GET /key/:key
	app.Get("/key/:key", handlers.GetKeyValue)

	// Start the server on port 3000
	log.Fatal(app.Listen(":3000"))
}
