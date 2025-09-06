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

	fmt.Printf("Starting replication handshake with master %s:%s\n",
		replState.MasterHost, replState.MasterPort)

	time.Sleep(100 * time.Millisecond)
	go performReplicationHandshake(replState.MasterHost, replState.MasterPort, serverPort)
}

func performReplicationHandshake(masterHost, masterPort, serverPort string) {
	masterAddr := fmt.Sprintf("%s:%s", masterHost, masterPort)
	conn, err := net.Dial("tcp", masterAddr)
	if err != nil {
		fmt.Printf("Failed to connect to master %s: %v\n", masterAddr, err)
		return
	}
	// CRITICAL FIX: Remove defer conn.Close() to keep connection alive!

	fmt.Printf("Connected to master %s\n", masterAddr)

	// Perform handshake
	if !sendCommand(conn, "*1\r\n$4\r\nPING\r\n") {
		fmt.Println("Failed to send PING to master")
		conn.Close()
		return
	}
	if !expectResponse(conn, "+PONG") {
		fmt.Println("Did not receive PONG from master")
		conn.Close()
		return
	}
	fmt.Println("✓ PING successful")

	replconfCmd := fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$%d\r\n%s\r\n",
		len(serverPort), serverPort)
	if !sendCommand(conn, replconfCmd) {
		fmt.Println("Failed to send REPLCONF listening-port to master")
		conn.Close()
		return
	}
	if !expectResponse(conn, "+OK") {
		fmt.Println("Did not receive OK for REPLCONF listening-port")
		conn.Close()
		return
	}
	fmt.Println("✓ REPLCONF listening-port successful")

	capaCmd := "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"
	if !sendCommand(conn, capaCmd) {
		fmt.Println("Failed to send REPLCONF capa to master")
		conn.Close()
		return
	}
	if !expectResponse(conn, "+OK") {
		fmt.Println("Did not receive OK for REPLCONF capa")
		conn.Close()
		return
	}
	fmt.Println("✓ REPLCONF capa successful")

	psyncCmd := "*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"
	if !sendCommand(conn, psyncCmd) {
		fmt.Println("Failed to send PSYNC to master")
		conn.Close()
		return
	}

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Failed to read PSYNC response: %v\n", err)
		conn.Close()
		return
	}
	response = strings.TrimSpace(response)

	if !strings.HasPrefix(response, "+FULLRESYNC") {
		fmt.Printf("Unexpected PSYNC response: %s\n", response)
		conn.Close()
		return
	}

	fmt.Printf("✓ PSYNC successful: %s\n", response)

	if !receiveRDB(reader) {
		fmt.Println("Failed to receive RDB file")
		conn.Close()
		return
	}

	fmt.Println("🎉 Replication handshake completed successfully!")

	// NEW: Keep connection alive and listen for propagated commands
	fmt.Println("📡 Starting to listen for propagated commands...")
	listenForPropagatedCommands(conn, reader)
}

// NEW FUNCTION: Listen for commands from master
func listenForPropagatedCommands(conn net.Conn, reader *bufio.Reader) {
	defer conn.Close()

	for {
		// Read command from master
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Connection to master lost: %v\n", err)
			return
		}

		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "*") {
			// Parse RESP Array from master
			numArgsStr := strings.TrimPrefix(line, "*")
			numArgs, err := strconv.Atoi(numArgsStr)
			if err != nil {
				fmt.Printf("Invalid RESP array from master: %s\n", line)
				continue
			}

			var parts []string
			for i := 0; i < numArgs; i++ {
				// Read bulk string header: $<length>
				_, err := reader.ReadString('\n')
				if err != nil {
					fmt.Printf("Failed to read bulk string header: %v\n", err)
					return
				}
				// Read actual content
				content, err := reader.ReadString('\n')
				if err != nil {
					fmt.Printf("Failed to read bulk string content: %v\n", err)
					return
				}
				parts = append(parts, strings.TrimSpace(content))
			}

			if len(parts) > 0 {
				fmt.Printf("📥 Received propagated command: %v\n", parts)
				// Process command locally without sending response
				processReplicatedCommand(parts)
			}
		}
	}
}

// NEW FUNCTION: Process commands received from master (no response sent)
func processReplicatedCommand(args []string) {
	if len(args) == 0 {
		return
	}

	fmt.Printf("Processing replicated command: %v\n", args)

	// Process command silently (no response to master)
	switch strings.ToUpper(args[0]) {
	case "SET":
		if len(args) >= 3 {
			key := args[1]
			value := args[2]
			store.Set(key, value, 0) // No TTL for simplicity
			fmt.Printf("✓ Replicated SET %s = %s\n", key, value)
		}
	case "DEL":
		if len(args) >= 2 {
			key := args[1]
			store.Delete(key)
			fmt.Printf("✓ Replicated DEL %s\n", key)
		}
	// Add more commands as needed
	default:
		fmt.Printf("Unknown replicated command: %s\n", args[0])
	}
}

func receiveRDB(reader *bufio.Reader) bool {
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Failed to read RDB bulk string header: %v\n", err)
		return false
	}

	line = strings.TrimSpace(line)

	if !strings.HasPrefix(line, "$") {
		fmt.Printf("Expected RDB bulk string, got: %s\n", line)
		return false
	}

	lengthStr := strings.TrimPrefix(line, "$")
	rdbLength, err := strconv.Atoi(lengthStr)
	if err != nil {
		fmt.Printf("Invalid RDB length: %s\n", lengthStr)
		return false
	}

	if rdbLength == -1 {
		fmt.Println("Received null RDB")
		return true
	}

	rdbData := make([]byte, rdbLength)
	bytesRead, err := io.ReadFull(reader, rdbData)
	if err != nil {
		fmt.Printf("Failed to read RDB data: %v\n", err)
		return false
	}

	if bytesRead != rdbLength {
		fmt.Printf("RDB length mismatch: expected %d, got %d\n", rdbLength, bytesRead)
		return false
	}

	fmt.Printf("✓ Received RDB file (%d bytes)\n", rdbLength)

	if validateRDB(rdbData) {
		fmt.Println("✓ RDB validation successful")
		return true
	} else {
		fmt.Println("✗ RDB validation failed")
		return false
	}
}

func validateRDB(rdbData []byte) bool {
	if len(rdbData) < 9 {
		return false
	}

	magic := string(rdbData[:5])
	if magic != "REDIS" {
		fmt.Printf("Invalid RDB magic: %s\n", magic)
		return false
	}

	version := string(rdbData[5:9])
	fmt.Printf("RDB version: %s\n", version)

	return true
}

func sendCommand(conn net.Conn, command string) bool {
	_, err := conn.Write([]byte(command))
	if err != nil {
		fmt.Printf("Failed to send command: %v\n", err)
		return false
	}
	return true
}

func expectResponse(conn net.Conn, expected string) bool {
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		return false
	}
	response = strings.TrimSpace(response)
	return response == expected
}
