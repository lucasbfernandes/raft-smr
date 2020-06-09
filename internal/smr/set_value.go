package smr

import (
	"encoding/json"
	"github.com/hashicorp/raft"
	"raft-smr/internal/fsm"
	"time"
)

type SetValueRequest struct {
	Key		string	`json:"key"`
	Value	string	`json:"value"`
}

func ExecuteSetValue(setValueRequest *SetValueRequest, raftInstance *raft.Raft) error {
	command := &fsm.Command{
		Op:    "set",
		Key:   setValueRequest.Key,
		Value: setValueRequest.Value,
	}

	buff, err := json.Marshal(command)
	if err != nil {
		return err
	}

	future := raftInstance.Apply(buff, 10 * time.Second)
	if err := future.Error(); err != nil {
		return err
	}

	return nil
}