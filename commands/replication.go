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

	_ = requestedReplID
	_ = requestedOffset

	response := fmt.Sprintf("+FULLRESYNC %s %d\r\n",
		replState.MasterReplID,
		replState.MasterReplOffset)
	conn.Write([]byte(response))

	// replicaAddr := conn.RemoteAddr().String()
	// store.AddReplica(replicaAddr)
	// fmt.Printf("Replica connected: %s\n", replicaAddr)

	store.AddReplicaWithConnection(conn)

	sendEmptyRDB(conn)
}

func sendEmptyRDB(conn net.Conn) {
	emptyRDB := generateEmptyRDB()

	rdbResponse := fmt.Sprintf("$%d\r\n", len(emptyRDB))
	fullResponse := append([]byte(rdbResponse), emptyRDB...)

	_, err := conn.Write(fullResponse)
	if err != nil {
		fmt.Printf("Failed to send RDB: %v\n", err)
		return
	}

	fmt.Printf("Sent empty RDB file (%d bytes) to replica\n", len(emptyRDB))
}

func generateEmptyRDB() []byte {
	rdb := []byte{
		0x52, 0x45, 0x44, 0x49, 0x53, // "REDIS"
		0x30, 0x30, 0x31, 0x31, // "0011" (version)
		0xfa, 0x09, 0x72, 0x65, 0x64, 0x69, 0x73, 0x2d, 0x76, 0x65, 0x72, // redis-ver
		0x05, 0x37, 0x2e, 0x32, 0x2e, 0x30, // "7.2.0"
		0xfa, 0x0a, 0x72, 0x65, 0x64, 0x69, 0x73, 0x2d, 0x62, 0x69, 0x74, 0x73, // redis-bits
		0xc0, 0x40, // 64
		0xfa, 0x05, 0x63, 0x74, 0x69, 0x6d, 0x65, // ctime
		0xc2, 0x6d, 0x08, 0xbc, 0x65, // timestamp
		0xfa, 0x08, 0x75, 0x73, 0x65, 0x64, 0x2d, 0x6d, 0x65, 0x6d, // used-mem
		0xc2, 0xb0, 0xc4, 0x10, 0x00, // memory value
		0xfa, 0x08, 0x61, 0x6f, 0x66, 0x2d, 0x62, 0x61, 0x73, 0x65, // aof-base
		0xc0, 0x00, // 0
		0xff,                                           // EOF
		0xf0, 0x6e, 0x3b, 0xfe, 0xc0, 0xff, 0x5a, 0xa2, // CRC64 checksum
	}
	return rdb
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
