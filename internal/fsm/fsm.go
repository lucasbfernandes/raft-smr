package fsm

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/raft"
	"io"
	"sync"
)

type FSM struct {
	mutex sync.Mutex
	mapData map[string]string
}

type Command struct {
	Op    string `json:"op,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type fsmSnapshot struct {
	store map[string]string
}

func CreateFSM() *FSM {
	return &FSM{
		mutex:  sync.Mutex{},
		mapData: make(map[string]string),
	}
}

func (fsm *FSM) Apply(l *raft.Log) interface{} {
	var c Command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}

	switch c.Op {
	case "set":
		return fsm.applySet(c.Key, c.Value)
	case "delete":
		return fsm.applyDelete(c.Key)
	case "get":
		return fsm.applyGet(c.Key)
	default:
		panic(fmt.Sprintf("unrecognized command op: %s", c.Op))
	}
}

func (fsm *FSM) Snapshot() (raft.FSMSnapshot, error) {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()

	object := make(map[string]string)
	for key, value := range fsm.mapData {
		object[key] = value
	}
	return &fsmSnapshot{ store: object }, nil
}

func (fsm *FSM) Restore(rc io.ReadCloser) error {
	object := make(map[string]string)
	if err := json.NewDecoder(rc).Decode(&object); err != nil {
		return err
	}

	fsm.mapData = object
	return nil
}

func (fsm *FSM) applySet(key, value string) interface{} {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()
	fsm.mapData[key] = value
	return nil
}

func (fsm *FSM) applyDelete(key string) interface{} {
	fsm.mutex.Lock()
	defer fsm.mutex.Unlock()
	delete(fsm.mapData, key)
	return nil
}

func (fsm *FSM) applyGet(key string) interface{} {
	return fsm.mapData[key]
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		b, err := json.Marshal(f.store)
		if err != nil {
			return err
		}

		if _, err := sink.Write(b); err != nil {
			return err
		}
		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

func (f *fsmSnapshot) Release() {

}