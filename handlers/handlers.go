package handlers

import (
	"ToyDistributedKeyValue/config"
	"ToyDistributedKeyValue/kvstore"
	toyraft "ToyDistributedKeyValue/raft"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/raft"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var raftNode *toyraft.RaftNode

func InitRaftNode(nodeID, addr string, fsm *kvstore.FSM, peers []string, bootstrapCluster bool) error {
	// Initialize the Raft node
	var err error
	raftNode, err = toyraft.NewRaftNode(nodeID, addr, fsm, peers, bootstrapCluster)
	if err != nil {
		return fmt.Errorf("failed to initialize Raft: %v", err)
	}
	return nil
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

	// Check whether the Raft node is a leader
	if raftNode.Raft.State() != raft.Leader {
		// If it is not a leader, redirect the request to the leader
		// If this node is not the leader, forward the request to the leader
		leaderAddr, _ := raftNode.Raft.LeaderWithID() // Returns Raft address of the leader
		if leaderAddr == "" {
			log.Println("No Raft leader available")
			return c.Status(http.StatusServiceUnavailable).SendString("No leader available")
		}

		// Forward the request to the leader
		resp, err := forwardToLeader(string(leaderAddr), key, value)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Failed to forward request to leader")
		}
		return c.Send(resp)
	} else {
		// It is the leader, so propose the new key:value to Raft
		if err := raftNode.ProposeSet(key, value); err != nil {
			return c.Status(http.StatusInternalServerError).SendString(fmt.Sprintf("Failed to set value: %v", err))
		}
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

	value, err := raftNode.Store.Get(key)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(fmt.Sprintf("Failed to get value: %v", err))
	}
	if value == "" {
		return c.Status(http.StatusNotFound).SendString("Key not found")
	}

	// Return the key-value pair
	return c.JSON(fiber.Map{
		"key":   key,
		"value": value,
	})
}

func HealthCheck(ctx *fiber.Ctx) error {
	// Check that the raftNode is not nil
	if raftNode == nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Raft node is not initialized",
		})
	}

	// Return success message
	return ctx.JSON(fiber.Map{
		"message": "Database is alive",
	})
}

// forwardToLeader forwards a POST request to the leader
func forwardToLeader(leaderAddr, key, value string) ([]byte, error) {
	// Make a POST request to the leader with the key and value as URL-encoded parameters

	// Take the host name from the leaderAddr only
	// The leaderAddr is in the format of "<host>:<port>"
	leaderHostname := strings.Split(leaderAddr, ":")[0]
	resp, err := http.Post(fmt.Sprintf("http://%s:%d/key?key=%s&value=%s", leaderHostname, config.HttpApiPort,
		url.QueryEscape(key),
		url.QueryEscape(value)), "application/x-www-form-urlencoded", nil)
	if err != nil {
		log.Printf("Failed to forward request to leader: %v", err)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return nil, err
	}
	return body, nil
}
