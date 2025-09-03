package commands

import (
	"fmt"
	"net"
	"strings"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handlePsync(args []string, conn net.Conn) {

	replState := store.GetReplicationState()

	if replState.Role != "master" {
		conn.Write([]byte("-ERR PSYNC can only be used on master\r\n"))
		return
	}

	if len(args) != 3 {
		conn.Write([]byte("-ERR wrong number of arguments for PSYNC\r\n"))
		return
	}

	requestedReplID := args[1]
	requestedOffset := args[2]

	//Full resync (simple implementation)

	_ = requestedReplID
	_ = requestedOffset

	response := fmt.Sprintf("+FULLRESYNC %s %d\r\n",
		replState.MasterReplID,
		replState.MasterReplOffset)
	conn.Write([]byte(response))

	replicaAddr := conn.RemoteAddr().String()
	store.AddReplica(replicaAddr)
	fmt.Printf("Replica connected: %s\n", replicaAddr)

	//TODO: After FULLRESYNC, we should send an empty RDB file

}

func handleReplconf(args []string, conn net.Conn) {
	if len(args) < 3 {
		conn.Write([]byte("-ERR wrong number of arguments for REPLCONF\r\n"))
		return
	}

	subcommand := strings.ToUpper(args[1])

	switch subcommand {
	case "LISTENING-PORT":
		if len(args) != 3 {
			conn.Write([]byte("-ERR wrong number of arguments for REPLCONF listening-port\r\n"))
			return
		}
		conn.Write([]byte("+OK\r\n"))

	case "CAPA":
		if len(args) != 3 {
			conn.Write([]byte("-ERR wrong number of arguments for REPLCONF capa\r\n"))
			return
		}

		conn.Write([]byte("+OK\r\n"))

	case "ACK":

		if len(args) != 3 {
			conn.Write([]byte("-ERR wrong number of arguments for REPLCONF ack\r\n"))
			return
		}

		conn.Write([]byte("+OK\r\n"))

	default:
		conn.Write([]byte("-ERR unknown REPLCONF option\r\n"))
	}

}
