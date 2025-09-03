package server

import (
	"bufio"
	"fmt"
	"net"
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

	//Starting handshake within a Goroutine so it does'nt block main server

	go performReplicationHandshake(replState.MasterHost, replState.MasterPort, serverPort)

}

func performReplicationHandshake(masterHost, masterPort, serverPort string) {

	//Connecting to master
	masterAddr := fmt.Sprintf("%s:%s", masterHost, masterPort)
	conn, err := net.Dial("tcp", masterAddr)

	if err != nil {
		fmt.Printf("Failed to connect to master %s: %v\n", masterAddr, err)
		return
	}

	defer conn.Close()

	fmt.Printf("Connected to master %s\n", masterAddr)

	//Sending ping(first step)

	if !sendCommand(conn, "*1\r\n$4\r\nPING\r\n") {
		fmt.Println("Failed to send PING to master")
		return
	}

	if !expectResponse(conn, "+PONG") {
		fmt.Println("Did not recieve PONG from master")
		return
	}

	fmt.Println("✓ PING successful")

	//Sending REPLCONF listening-port
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

	//Sending REPLCONF capa PSYNC2
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

	//Sending full psync
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

	fmt.Println("🎉 Replication handshake completed successfully!")
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
