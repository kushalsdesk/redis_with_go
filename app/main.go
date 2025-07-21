package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// Global Level

var store = make(map[string]string)
var exprMap = make(map[string]time.Time)

func main() {
	fmt.Println("Listening on 0.0.0.0:6379...")
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		// Read the array header: *2\r\n
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected.")
			return
		}

		line = strings.TrimSpace(line)

		if !strings.HasPrefix(line, "*") {
			fmt.Println("Invalid RESP message:", line)
			return
		}

		// Read command part ($4\r\nECHO\r\n)
		_, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading bulk string header")
			return
		}

		// Read: PING\r\n
		cmdLine, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading command line")
			return
		}

		cmd := strings.ToUpper(strings.TrimSpace(cmdLine))

		switch cmd {
		case "PING":
			conn.Write([]byte("+PONG\r\n"))

		case "ECHO":
			_, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			argline, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			arg := strings.TrimSpace(argline)
			// Format as RESP bult string
			resp := fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
			conn.Write([]byte(resp))

		case "SET":
			// READ Key
			_, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			keyline, err := reader.ReadString('\n')
			if err != nil {
				return
			}

			key := strings.TrimSpace(keyline)

			// READ values
			_, error := reader.ReadString('\n')
			if error != nil {
				return
			}
			valueline, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			value := strings.TrimSpace(valueline)

			store[key] = value

			conn.Write([]byte("+OK\r\n"))

		case "GET":
			_, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			keyline, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			key := strings.TrimSpace(keyline)

			//  lookup in map
			value, ok := store[key]
			if !ok {
				// return null bulk string
				conn.Write([]byte("$-1\r\n"))
			} else {
				resp := fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
				conn.Write([]byte(resp))
			}

		default:
			conn.Write([]byte("-ERR unknown command\r\n"))

		}
	}
}
