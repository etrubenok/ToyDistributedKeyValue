package main

import (
	"ToyDistributedKeyValue/config"
	"ToyDistributedKeyValue/handlers"
	"ToyDistributedKeyValue/kvstore"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"os"
	"strings"
)

func main() {
	// Load configuration (e.g., node ID, address, peers, etc.)
	nodeID := os.Getenv("NODE_ID")
	addr := os.Getenv("NODE_ADDR")
	peersStr := os.Getenv("RAFT_PEERS")
	bootstrap := os.Getenv("RAFT_BOOTSTRAP")

	// Convert the bootstrap environment variable to a boolean
	bootstrapCluster := bootstrap == "true"

	var peers []string
	// Split the peers string into a slice
	if peersStr != "" {
		peers = strings.Split(peersStr, ",")
	}

	// Panic if any of the required environment variables are missing
	if nodeID == "" || addr == "" || peersStr == "" {
		log.Fatalf("Missing required environment variables (NODE_ID, NODE_ADDR, RAFT_PEERS)")
	}

	// Log the configuration
	log.Printf("Node ID: %s", nodeID)
	log.Printf("Node Address: %s", addr)
	log.Printf("Raft Peers: %v", peers)
	log.Printf("Bootstrap Cluster: %v", bootstrapCluster)

	// Create a new directory for the node's data
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize the key-value store backed by RocksDB
	store, err := kvstore.NewKVStore("./data")
	if err != nil {
		log.Fatalf("Failed to initialize KV store: %v", err)
	}

	// Initialize Raft FSM using the KV store
	fsm := kvstore.NewFSM(store)

	// Initialize the Raft node
	err = handlers.InitRaftNode(nodeID, addr, fsm, peers, bootstrapCluster)
	if err != nil {
		log.Fatalf("Failed to initialize Raft Node: %v", err)
	}

	// Initialize a new Fiber app
	app := fiber.New()

	// Route to health check
	app.Get("/healthcheck", handlers.HealthCheck)

	// Route to handle POST /key
	app.Post("/key", handlers.SetKeyValue)

	// Route to handle GET /key/:key
	app.Get("/key/:key", handlers.GetKeyValue)

	// Start the server on port HTTP_API_PORT
	log.Fatal(app.Listen(fmt.Sprintf(":%d", config.HttpApiPort)))
}
