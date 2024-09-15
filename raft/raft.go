package raft

import (
	"ToyDistributedKeyValue/kvstore"
	"fmt"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"log"
	"os"
	"time"
)

const raftTimeout = 10 * time.Second

type RaftNode struct {
	Raft  *raft.Raft
	Store *kvstore.KVStore
}

// NewRaftNode initializes a new Raft node
func NewRaftNode(nodeID, addr string, fsm *kvstore.FSM, peers []string, bootstrapCluster bool) (*RaftNode, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	// Setup BoltDB as the log store for Raft
	logStore, err := raftboltdb.NewBoltStore("./raft-log-" + nodeID + ".db")
	if err != nil {
		return nil, err
	}

	// Use in-memory snapshot store (could be file-based for persistence)
	snapshotStore := raft.NewInmemSnapshotStore()

	// Create the transport using the specified Raft address
	transport, err := raft.NewTCPTransport(addr, nil, 3, raftTimeout, os.Stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %v", err)
	}

	// Initialize the Raft system
	raftSystem, err := raft.NewRaft(config, fsm, logStore, logStore, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	configFuture := raftSystem.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		log.Fatalf("failed to get Raft configuration: %v", err)
	}

	// Check if there are any servers in the configuration
	if len(configFuture.Configuration().Servers) == 0 {
		// One-time bootstrap of the Raft cluster if we are starting the first node
		if bootstrapCluster {
			log.Println("Bootstrapping new Raft cluster")
			// Combine to nodeID/addr with peers into Servers slice
			servers := []raft.Server{
				{
					ID:      raft.ServerID(nodeID),
					Address: raft.ServerAddress(addr),
				},
			}

			// Add peers to the Servers slice
			for _, peer := range peers {
				servers = append(servers, raft.Server{
					ID:      raft.ServerID(peer),
					Address: raft.ServerAddress(peer),
				})
			}

			err = raftSystem.BootstrapCluster(raft.Configuration{
				Servers: servers,
			}).Error()
			if err != nil {
				return nil, fmt.Errorf("failed to bootstrap cluster: %v", err)
			}
		}
	}

	// Add peers to the Raft cluster
	if !bootstrapCluster {
		for _, peer := range peers {
			log.Println("Adding peer:", peer)
			raftSystem.AddVoter(raft.ServerID(peer), raft.ServerAddress(peer), 0, time.Second)
		}
	}

	return &RaftNode{Raft: raftSystem, Store: fsm.Store()}, nil
}

// ProposeSet proposes a key-value set operation to the Raft cluster
func (r *RaftNode) ProposeSet(key, value string) error {
	command := kvstore.Command{
		Operation: "SET",
		Key:       key,
		Value:     value,
	}
	data, err := command.Marshal()
	if err != nil {
		return err
	}
	return r.Raft.Apply(data, raftTimeout).Error()
}
