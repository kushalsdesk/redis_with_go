package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/kushalsdesk/redis_with_go/store"
)

func StartReplicationClient(serverPort string) {
	replState := store.GetReplicationState()
	if replState.Role != "slave" {
		return
	}

	fmt.Printf("🚀 Starting replication with master %s:%s\n", replState.MasterHost, replState.MasterPort)
	time.Sleep(100 * time.Millisecond)
	go performReplicationHandshake(replState.MasterHost, replState.MasterPort, serverPort)
}

func performReplicationHandshake(masterHost, masterPort, serverPort string) {
	masterAddr := fmt.Sprintf("%s:%s", masterHost, masterPort)
	conn, err := net.Dial("tcp", masterAddr)
	if err != nil {
		fmt.Printf("❌ Failed to connect to master %s: %v\n", masterAddr, err)
		return
	}

	fmt.Printf("🔗 Connected to master %s\n", masterAddr)

	if !performHandshakeSteps(conn, serverPort) {
		conn.Close()
		return
	}

	fmt.Printf("🎉 Replication handshake completed!\n")
	fmt.Printf("📡 Listening for propagated commands...\n")

	reader := bufio.NewReader(conn)
	listenForPropagatedCommands(conn, reader)
}

func performHandshakeSteps(conn net.Conn, serverPort string) bool {
	// Step 1: PING
	if !sendCommand(conn, "*1\r\n$4\r\nPING\r\n") ||
		!expectResponse(conn, "+PONG") {
		fmt.Printf("❌ PING handshake failed\n")
		return false
	}
	fmt.Printf("✅ PING successful\n")

	// Step 2: REPLCONF listening-port
	replconfCmd := fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$%d\r\n%s\r\n",
		len(serverPort), serverPort)
	if !sendCommand(conn, replconfCmd) ||
		!expectResponse(conn, "+OK") {
		fmt.Printf("❌ REPLCONF listening-port failed\n")
		return false
	}
	fmt.Printf("✅ REPLCONF listening-port successful\n")

	// Step 3: REPLCONF capa
	capaCmd := "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"
	if !sendCommand(conn, capaCmd) ||
		!expectResponse(conn, "+OK") {
		fmt.Printf("❌ REPLCONF capa failed\n")
		return false
	}
	fmt.Printf("✅ REPLCONF capa successful\n")

	// Step 4: PSYNC
	psyncCmd := "*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"
	if !sendCommand(conn, psyncCmd) {
		fmt.Printf("❌ PSYNC send failed\n")
		return false
	}

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("❌ PSYNC response read failed: %v\n", err)
		return false
	}

	response = strings.TrimSpace(response)
	if !strings.HasPrefix(response, "+FULLRESYNC") {
		fmt.Printf("❌ Unexpected PSYNC response: %s\n", response)
		return false
	}
	fmt.Printf("✅ PSYNC successful: %s\n", response)

	if !receiveRDB(reader) {
		fmt.Printf("❌ RDB receive failed\n")
		return false
	}

	return true
}

func listenForPropagatedCommands(conn net.Conn, reader *bufio.Reader) {
	defer conn.Close()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("📡 Connection to master lost: %v\n", err)
			return
		}

		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "*") {
			if !processRESPArray(reader, line) {
				continue
			}
		}
	}
}

func processRESPArray(reader *bufio.Reader, line string) bool {
	numArgsStr := strings.TrimPrefix(line, "*")
	numArgs, err := strconv.Atoi(numArgsStr)
	if err != nil {
		fmt.Printf("❌ Invalid RESP array: %s\n", line)
		return false
	}

	var parts []string
	for i := 0; i < numArgs; i++ {
		// Read bulk string header
		_, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("❌ Failed to read bulk string header: %v\n", err)
			return false
		}

		content, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("❌ Failed to read bulk string content: %v\n", err)
			return false
		}
		parts = append(parts, strings.TrimSpace(content))
	}

	if len(parts) > 0 {
		fmt.Printf("📥 Received: %v\n", parts)
		processReplicatedCommand(parts)
	}
	return true
}

func processReplicatedCommand(args []string) {
	if len(args) == 0 {
		return
	}

	command := strings.ToUpper(args[0])
	switch command {
	case "SET":
		if len(args) >= 3 {
			key, value := args[1], args[2]
			store.Set(key, value, 0)
			fmt.Printf("✅ Replicated SET %s = %s\n", key, value)
		}
	case "DEL":
		if len(args) >= 2 {
			key := args[1]
			store.Delete(key)
			fmt.Printf("✅ Replicated DEL %s\n", key)
		}
	case "LPUSH":
		if len(args) >= 3 {
			key := args[1]
			elements := args[2:]
			length := store.ListPush(key, elements, true)
			fmt.Printf("✅ Replicated LPUSH %s (length: %d)\n", key, length)
		}
	case "RPUSH":
		if len(args) >= 3 {
			key := args[1]
			elements := args[2:]
			length := store.ListPush(key, elements, false)
			fmt.Printf("✅ Replicated RPUSH %s (length: %d)\n", key, length)
		}
	case "LPOP":
		if len(args) >= 2 {
			key := args[1]
			count := 1
			if len(args) == 3 {
				if parsedCount, err := strconv.Atoi(args[2]); err == nil && parsedCount >= 0 {
					count = parsedCount
				} else {
					fmt.Printf("❌ Invalid LPOP count: %s\n", args[2])
					return
				}
			}
			elements, exists := store.ListPopMultiple(key, count, true)
			if !exists {
				fmt.Printf("✅ Replicated LPOP %s (key not found or wrong type)\n", key)
			} else {
				fmt.Printf("✅ Replicated LPOP %s (popped %d elements)\n", key, len(elements))
			}
		}
	case "RPOP":
		if len(args) >= 2 {
			key := args[1]
			count := 1
			if len(args) == 3 {
				if parsedCount, err := strconv.Atoi(args[2]); err == nil && parsedCount >= 0 {
					count = parsedCount
				} else {
					fmt.Printf("❌ Invalid RPOP count: %s\n", args[2])
					return
				}
			}
			elements, exists := store.ListPopMultiple(key, count, false)
			if !exists {
				fmt.Printf("✅ Replicated RPOP %s (key not found or wrong type)\n", key)
			} else {
				fmt.Printf("✅ Replicated RPOP %s (popped %d elements)\n", key, len(elements))
			}
		}
	case "INCR":
		if len(args) >= 2 {
			key := args[1]
			newValue, err := store.Increment(key)
			if err != nil {
				fmt.Printf("❌ Failed to replicate INCR %s: %v\n", key, err)
			} else {
				fmt.Printf("✅ Replicated INCR %s = %d\n", key, newValue)
			}
		}
	case "DECR":
		if len(args) >= 2 {
			key := args[1]
			newValue, err := store.Decrement(key)
			if err != nil {
				fmt.Printf("❌ Failed to replicate DECR %s: %v\n", key, err)
			} else {
				fmt.Printf("✅ Replicated DECR %s = %d\n", key, newValue)
			}
		}
	case "INCRBY":
		if len(args) == 2 {
			key := args[1]
			newValue, err := store.Increment(key)
			if err != nil {
				fmt.Printf("❌ Failed to replicate INCRBY %s: %v\n", key, err)
			} else {
				fmt.Printf("✅ Replicated INCRBY %s (default 1) = %d\n", key, newValue)
			}
		} else if len(args) >= 3 {
			key := args[1]
			amount, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				fmt.Printf("❌ Invalid INCRBY amount: %s\n", args[2])
				return
			}
			newValue, err := store.IncrementBy(key, amount)
			if err != nil {
				fmt.Printf("❌ Failed to replicate INCRBY %s %d: %v\n", key, amount, err)
			} else {
				fmt.Printf("✅ Replicated INCRBY %s %d = %d\n", key, amount, newValue)
			}
		}
	case "DECRBY":
		if len(args) == 2 {
			key := args[1]
			newValue, err := store.Decrement(key)
			if err != nil {
				fmt.Printf("❌ Failed to replicate DECRBY %s: %v\n", key, err)
			} else {
				fmt.Printf("✅ Replicated DECRBY %s (default 1) = %d\n", key, newValue)
			}
		} else if len(args) >= 3 {
			key := args[1]
			amount, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				fmt.Printf("❌ Invalid DECRBY amount: %s\n", args[2])
				return
			}
			newValue, err := store.DecrementBy(key, amount)
			if err != nil {
				fmt.Printf("❌ Failed to replicate DECRBY %s %d: %v\n", key, amount, err)
			} else {
				fmt.Printf("✅ Replicated DECRBY %s %d = %d\n", key, amount, newValue)
			}
		}
	default:
		fmt.Printf("⚠️ Unknown replicated command: %s\n", command)
	}
}

func receiveRDB(reader *bufio.Reader) bool {
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("❌ Failed to read RDB header: %v\n", err)
		return false
	}

	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "$") {
		fmt.Printf("❌ Expected RDB bulk string, got: %s\n", line)
		return false
	}

	lengthStr := strings.TrimPrefix(line, "$")
	rdbLength, err := strconv.Atoi(lengthStr)
	if err != nil {
		fmt.Printf("❌ Invalid RDB length: %s\n", lengthStr)
		return false
	}

	if rdbLength == -1 {
		fmt.Printf("📦 Received null RDB\n")
		return true
	}

	rdbData := make([]byte, rdbLength)
	bytesRead, err := io.ReadFull(reader, rdbData)
	if err != nil {
		fmt.Printf("❌ Failed to read RDB data: %v\n", err)
		return false
	}

	if bytesRead != rdbLength {
		fmt.Printf("❌ RDB length mismatch: expected %d, got %d\n", rdbLength, bytesRead)
		return false
	}

	fmt.Printf("📦 Received RDB file (%d bytes)\n", rdbLength)

	if validateRDB(rdbData) {
		fmt.Printf("✅ RDB validation successful\n")
		return true
	}

	fmt.Printf("❌ RDB validation failed\n")
	return false
}

func validateRDB(rdbData []byte) bool {
	if len(rdbData) < 9 {
		return false
	}

	magic := string(rdbData[:5])
	if magic != "REDIS" {
		fmt.Printf("❌ Invalid RDB magic: %s\n", magic)
		return false
	}

	version := string(rdbData[5:9])
	fmt.Printf("📋 RDB version: %s\n", version)
	return true
}

func sendCommand(conn net.Conn, command string) bool {
	_, err := conn.Write([]byte(command))
	if err != nil {
		fmt.Printf("❌ Failed to send command: %v\n", err)
		return false
	}
	return true
}

func expectResponse(conn net.Conn, expected string) bool {
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		return false
	}

	response = strings.TrimSpace(response)
	return response == expected
}
