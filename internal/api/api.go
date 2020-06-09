package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"net"
	"os"
	"raft-smr/internal/configuration"
	"raft-smr/internal/controllers"
	"raft-smr/internal/fsm"
	"time"
)

func StartRaft(nodeID string, raftAddress string, raftDir string) error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	address, err := net.ResolveTCPAddr("tcp", raftAddress)
	if err != nil {
		return err
	}

	transport, err := raft.NewTCPTransport(raftAddress, address, 3, 10 * time.Second, os.Stderr)
	if err != nil {
		return err
	}

	snapshots, err := raft.NewFileSnapshotStore(raftDir, 2, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	ra, err := raft.NewRaft(config, fsm.CreateFSM(), logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}

	clusterConfig := configuration.GetConfiguration()
	var servers []raft.Server
	for _, member := range clusterConfig.Members {
		servers = append(servers, raft.Server{
			ID:       raft.ServerID(member.NodeID),
			Address:  raft.ServerAddress(member.RaftAddress),
		})
	}
	raftConfig := raft.Configuration{ Servers: servers }
	ra.BootstrapCluster(raftConfig)

	return nil
}

func StartAPI(port string) {
	router := gin.Default()
	router.GET("/set", controllers.SetValue)
	router.GET("/get", controllers.GetValue)
	router.Run(":" + port)
}