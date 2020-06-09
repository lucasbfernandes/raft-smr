package smr

import (
	"encoding/json"
	"github.com/hashicorp/raft"
	"raft-smr/internal/fsm"
	"time"
)

type GetValueRequest struct {
	Key    string `form:"key"`
}

func ExecuteGetValue(getValueRequest *GetValueRequest, raftInstance *raft.Raft) (string, error) {
	command := &fsm.Command{
		Op:    "get",
		Key:   getValueRequest.Key,
	}

	buff, err := json.Marshal(command)
	if err != nil {
		return "", err
	}

	future := raftInstance.Apply(buff, 10 * time.Second)
	if err := future.Error(); err != nil {
		return "", err
	}

	return future.Response().(string), nil
}
