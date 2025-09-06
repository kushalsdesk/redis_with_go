package commands

import (
	"fmt"
	"strings"

	"github.com/kushalsdesk/redis_with_go/store"
)

// Define which commands are write commands that should be propagated
var writeCommands = map[string]bool{
	"SET":    true,
	"DEL":    true,
	"LPUSH":  true,
	"RPUSH":  true,
	"LPOP":   true,
	"RPOP":   true,
	"INCR":   true,
	"INCRBY": true,
	"DECR":   true,
	"DECRBY": true,
	"XADD":   true,
}

func IsWriteCommand(command string) bool {
	return writeCommands[strings.ToUpper(command)]
}

func EncodeRESPArray(args []string) []byte {
	if len(args) == 0 {
		return []byte("*0\r\n")
	}

	var result strings.Builder

	result.WriteString(fmt.Sprintf("*%d\r\n", len(args)))

	for _, arg := range args {
		result.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}

	return []byte(result.String())
}

func PropagateCommand(args []string) {
	if len(args) == 0 {
		return
	}

	if !IsWriteCommand(args[0]) {
		return
	}

	replState := store.GetReplicationState()
	if replState.Role != "master" {
		return
	}

	replicas := store.GetReplicaConnections()
	if len(replicas) == 0 {
		return
	}

	respCommand := EncodeRESPArray(args)

	fmt.Printf("Propagating command to %d replicas: %v\n", len(replicas), args)

	for _, replica := range replicas {
		go func(r *store.ReplicationConnection) {
			_, err := r.Connection.Write(respCommand)
			if err != nil {
				fmt.Printf("Failed to propagate command to replica %s: %v\n",
					r.Address, err)
				store.RemoveReplicaByConnection(r.Connection)
			}
		}(replica)
	}

	fmt.Printf("Command propagated successfully\n")
}
