package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kushalsdesk/redis_with_go/server"
)

func main() {

	port := flag.String("port", "6379", "Port to listen on")
	replicaof := flag.String("replicaof", "", "Master host and port")

	flag.Parse()

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

	server.ListenAndServe(addr)
}
