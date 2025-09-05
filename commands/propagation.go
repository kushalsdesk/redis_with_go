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
	// Add more write commands as needed
}

// Check if a command should be propagated to replicas
func IsWriteCommand(command string) bool {
	return writeCommands[strings.ToUpper(command)]
}

// Convert command arguments to RESP array format
func EncodeRESPArray(args []string) []byte {
	if len(args) == 0 {
		return []byte("*0\r\n")
	}

	var result strings.Builder

	// Write array header
	result.WriteString(fmt.Sprintf("*%d\r\n", len(args)))

	// Write each argument as bulk string
	for _, arg := range args {
		result.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}

	return []byte(result.String())
}

// Propagate command to all connected replicas
func PropagateCommand(args []string) {
	if len(args) == 0 {
		return
	}

	// Only propagate write commands
	if !IsWriteCommand(args[0]) {
		return
	}

	// Get current replication state
	replState := store.GetReplicationState()
	if replState.Role != "master" {
		return // Only masters propagate commands
	}

	// Get all active replica connections
	replicas := store.GetReplicaConnections()
	if len(replicas) == 0 {
		return // No replicas to propagate to
	}

	// Encode command as RESP array
	respCommand := EncodeRESPArray(args)

	fmt.Printf("Propagating command to %d replicas: %v\n", len(replicas), args)

	// Send to all replicas
	for _, replica := range replicas {
		go func(r *store.ReplicationConnection) {
			_, err := r.Connection.Write(respCommand)
			if err != nil {
				fmt.Printf("Failed to propagate command to replica %s: %v\n",
					r.Address, err)
				// Mark replica as disconnected
				store.RemoveReplicaByConnection(r.Connection)
			}
		}(replica)
	}

	fmt.Printf("Command propagated successfully\n")
}
