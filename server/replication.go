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
	defer conn.Close()

	fmt.Printf("Connected to master %s\n", masterAddr)

	if !sendCommand(conn, "*1\r\n$4\r\nPING\r\n") {
		fmt.Println("Failed to send PING to master")
		return
	}
	if !expectResponse(conn, "+PONG") {
		fmt.Println("Did not receive PONG from master")
		return
	}
	fmt.Println("✓ PING successful")

	replconfCmd := fmt.Sprintf("*3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$%d\r\n%s\r\n",
		len(serverPort), serverPort)
	if !sendCommand(conn, replconfCmd) {
		fmt.Println("Failed to send REPLCONF listening-port to master")
		return
	}
	if !expectResponse(conn, "+OK") {
		fmt.Println("Did not receive OK for REPLCONF listening-port")
		return
	}
	fmt.Println("✓ REPLCONF listening-port successful")

	capaCmd := "*3\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n"
	if !sendCommand(conn, capaCmd) {
		fmt.Println("Failed to send REPLCONF capa to master")
		return
	}
	if !expectResponse(conn, "+OK") {
		fmt.Println("Did not receive OK for REPLCONF capa")
		return
	}
	fmt.Println("✓ REPLCONF capa successful")

	psyncCmd := "*3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n"
	if !sendCommand(conn, psyncCmd) {
		fmt.Println("Failed to send PSYNC to master")
		return
	}

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Failed to read PSYNC response: %v\n", err)
		return
	}
	response = strings.TrimSpace(response)

	if !strings.HasPrefix(response, "+FULLRESYNC") {
		fmt.Printf("Unexpected PSYNC response: %s\n", response)
		return
	}

	fmt.Printf("✓ PSYNC successful: %s\n", response)

	if !receiveRDB(reader) {
		fmt.Println("Failed to receive RDB file")
		return
	}

	fmt.Println("🎉 Replication handshake completed successfully!")
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
