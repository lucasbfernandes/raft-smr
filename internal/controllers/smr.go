package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"raft-smr/internal/smr"
)

func SetValue(context *gin.Context, raftInstance *raft.Raft) {
	var setValueRequest *smr.SetValueRequest
	context.Bind(&setValueRequest)

	err := smr.ExecuteSetValue(setValueRequest, raftInstance)
	if err != nil {
		fmt.Println(err)
		context.JSON(500, gin.H{ "error": "Could not set value" })
	}
	context.JSON(201, nil)
}

func GetValue(context *gin.Context, raftInstance *raft.Raft) {
	var getValueRequest *smr.GetValueRequest
	context.Bind(&getValueRequest)

	value, err := smr.ExecuteGetValue(getValueRequest, raftInstance)
	if err != nil {
		fmt.Println(err)
		context.JSON(500, gin.H{ "error": "Could not get value" })
	}
	context.JSON(200, value)
}