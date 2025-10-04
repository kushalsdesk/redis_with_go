package commands

import (
	"fmt"
	"strings"

	"github.com/kushalsdesk/redis_with_go/store"
)

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
	if len(args) == 0 || !IsWriteCommand(args[0]) {
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
	cmdSize := store.EstimateCommandSize(args)
	fmt.Printf("📡 Propagating to %d replicas: %v (size ~%d bytes)\n", len(replicas), args, cmdSize)

	successCount := 0
	for _, replica := range replicas {
		go func(r *store.ReplicationConnection) {
			_, err := r.Connection.Write(respCommand)
			if err != nil {
				fmt.Printf("❌ Propagation failed to %s: %v\n", r.Address, err)
				store.RemoveReplicaByConnection(r.Connection)
			} else {
				successCount++
			}
		}(replica)
	}

	store.UpdateMasterOffset(cmdSize)
	fmt.Printf("✅ Command propagated successfully\n")
}
