package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/linxGnu/grocksdb"
	"log"
)

var db *grocksdb.DB

// InitDB Initialize the RocksDB database
func InitDB() {
	bbto := grocksdb.NewDefaultBlockBasedTableOptions()
	bbto.SetBlockCache(grocksdb.NewLRUCache(3 << 30))

	// Options for RocksDB
	opts := grocksdb.NewDefaultOptions()
	opts.SetBlockBasedTableFactory(bbto)
	opts.SetCreateIfMissing(true)

	// Open RocksDB database (this creates or opens the DB in "./testdb" directory)
	var err error
	db, err = grocksdb.OpenDb(opts, "./testdb")
	if err != nil {
		log.Fatalf("Failed to open RocksDB: %v", err)
	}
}

// CloseDB Close the RocksDB database
func CloseDB() {
	db.Close()
}

// SetKeyValue handles POST /key requests to set a key-value pair
func SetKeyValue(c *fiber.Ctx) error {
	// Get key and value from the URL-encoded form
	key := c.FormValue("key")
	value := c.FormValue("value")

	// Validate that both key and value are provided
	if key == "" || value == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Key and value are required",
		})
	}

	// Store the key-value pair in RocksDB
	wo := grocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()

	err := db.Put(wo, []byte(key), []byte(value))
	if err != nil {
		log.Printf("Failed to set key-value pair: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to set key-value pair",
		})
	}
	// Return success message
	return c.JSON(fiber.Map{
		"message": "Key-value pair set successfully",
	})
}

// GetKeyValue handles GET /key/:key requests to retrieve a value by key
func GetKeyValue(c *fiber.Ctx) error {
	// Extract the key from the URL parameter
	key := c.Params("key")

	// Create ReadOptions
	ro := grocksdb.NewDefaultReadOptions()
	defer ro.Destroy()

	// Get the value from RocksDB
	value, err := db.Get(ro, []byte(key))
	if err != nil {
		log.Printf("Failed to get value for key %s: %v", key, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get value",
		})
	}

	if !value.Exists() {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Key not found",
		})
	}

	// Return the key-value pair
	return c.JSON(fiber.Map{
		"key":   key,
		"value": string(value.Data()),
	})
}

func HealthCheck(ctx *fiber.Ctx) error {
	// Check that the database is open
	if db == nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database is not open",
		})
	}
	// Check that the database is alive
	status := db.GetProperty("rocksdb.stats")
	if status == "" {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database is not alive",
		})
	}
	// Return success message
	return ctx.JSON(fiber.Map{
		"message": "Database is alive",
	})
}
