package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kushalsdesk/redis_with_go/server"
	"github.com/kushalsdesk/redis_with_go/store"
)

func main() {
	port := flag.String("port", "6379", "Port to listen on")
	replicaof := flag.String("replicaof", "", "Master host and port")
	flag.Parse()

	// global port for replication handshake
	serverPort := *port

	if *replicaof != "" {
		parts := strings.Fields(*replicaof)
		if len(parts) != 2 {
			fmt.Println("ERR: --replicaof must be in format 'host port'")
			os.Exit(1)
		}
		masterHost := parts[0]
		masterPort := parts[1]
		store.SetReplicationRole("slave", masterHost, masterPort)
	}

	if *port == "" {
		fmt.Println("Port cannot be empty")
		os.Exit(1)
	}

	addr := fmt.Sprintf("0.0.0.0:%s", *port)
	if *replicaof != "" {
		fmt.Printf("Starting Redis server as replica of %s on %s ", *replicaof, addr)
	} else {
		fmt.Printf("Starting Redis server as master on %s", addr)
	}

	if *replicaof != "" {
		go func() {
			server.StartReplicationClient(serverPort)
		}()
	}

	server.ListenAndServe(addr)
}
