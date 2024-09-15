package kvstore

import (
	"encoding/json"
	"github.com/hashicorp/raft"
	"io"
)

type Command struct {
	Operation string
	Key       string
	Value     string
}

func (c *Command) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

func (c *Command) Unmarshal(data []byte) error {
	return json.Unmarshal(data, c)
}

type FSM struct {
	store *KVStore
}

func (fsm *FSM) Snapshot() (raft.FSMSnapshot, error) {
	//TODO implement me
	panic("implement me")
}

func (fsm *FSM) Restore(snapshot io.ReadCloser) error {
	//TODO implement me
	panic("implement me")
}

// NewFSM initializes a new FSM with the given key-value store
func NewFSM(store *KVStore) *FSM {
	return &FSM{store: store}
}

// Apply is called by Raft to apply a log entry to the FSM (state machine)
func (fsm *FSM) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := cmd.Unmarshal(log.Data); err != nil {
		return err
	}

	switch cmd.Operation {
	case "SET":
		return fsm.store.Set(cmd.Key, cmd.Value)
	}
	return nil
}

func (fsm *FSM) Store() *KVStore {
	return fsm.store
}
