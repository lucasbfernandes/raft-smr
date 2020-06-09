package main

import (
	"flag"
	"raft-smr/internal/api"
)

var (
	port = flag.String("port", "8080", "Http port")
	raftAddress = flag.String("raft", "127.0.0.1:9090", "Raft address")
	nodeID = flag.String("id", "node-1", "Node id")
	raftDir = flag.String("dir", "./data", "Raft dir")
)

func init() {
	flag.Parse()
}

func main() {
	api.StartRaft(*nodeID, *raftAddress, *raftDir)
	api.StartAPI(*port)
}
