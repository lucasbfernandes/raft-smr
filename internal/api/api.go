package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"net"
	"net/http/httputil"
	"net/url"
	"os"
	"raft-smr/internal/configuration"
	"raft-smr/internal/controllers"
	"raft-smr/internal/fsm"
	"time"
)

func StartRaft(nodeID string, raftAddress string, raftDir string, clusterConfig *configuration.Configuration) (*raft.Raft, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	address, err := net.ResolveTCPAddr("tcp", raftAddress)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransport(raftAddress, address, 3, 10 * time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}

	snapshots, err := raft.NewFileSnapshotStore(raftDir, 2, os.Stderr)
	if err != nil {
		fmt.Errorf("file snapshot store: %s", err)
		return nil, err
	}

	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	raftInstance, err := raft.NewRaft(config, fsm.CreateFSM(), logStore, stableStore, snapshots, transport)
	if err != nil {
		fmt.Errorf("new raft: %s", err)
		return nil, err
	}

	var servers []raft.Server
	for _, member := range clusterConfig.Members {
		servers = append(servers, raft.Server{
			ID:       raft.ServerID(member.NodeID),
			Address:  raft.ServerAddress(member.RaftAddress),
		})
	}
	raftConfig := raft.Configuration{ Servers: servers }
	raftInstance.BootstrapCluster(raftConfig)

	return raftInstance, nil
}

func StartAPI(port string, raftInstance *raft.Raft, clusterConfig *configuration.Configuration) {
	router := gin.Default()

	router.POST("/set", func(context *gin.Context) {
		if raftInstance.State() != raft.Leader {
			err := forwardToLeader(raftInstance, context, clusterConfig)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
		controllers.SetValue(context, raftInstance)
	})

	router.GET("/get", func(context *gin.Context) {
		if raftInstance.State() != raft.Leader {
			err := forwardToLeader(raftInstance, context, clusterConfig)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
		controllers.GetValue(context, raftInstance)
	})

	router.Run(":" + port)
}

func forwardToLeader(raftInstance *raft.Raft, context *gin.Context, clusterConfig *configuration.Configuration) error  {
	var address string
	for _, member := range clusterConfig.Members {
		if raft.ServerAddress(member.RaftAddress) == raftInstance.Leader() {
			address = member.HttpAddress
		}
	}

	leaderURL, err := url.Parse(address)
	if err != nil {
		return err
	}

	fmt.Printf("Forwarding request to leader on address %s", leaderURL)
	proxy := httputil.NewSingleHostReverseProxy(leaderURL)
	proxy.ServeHTTP(context.Writer, context.Request)
	return nil
}