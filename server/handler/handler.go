package handler

import (
	"bufio"
	"net"
	"strconv"
	"strings"

	"github.com/kushalsdesk/redis_with_go/commands"
	"github.com/kushalsdesk/redis_with_go/store"
)

func HandleConnection(conn net.Conn) {
	defer func() {
		store.RemoveReplicaByConnection(conn)
		conn.Close()
	}()

	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "*") {
			// Parse RESP array format
			parts := parseRESPArray(reader, line)
			if len(parts) > 0 {
				commands.Dispatch(parts, conn)
			}
		} else {
			// Parse simple string format
			parts := strings.Fields(line)
			if len(parts) > 0 {
				commands.Dispatch(parts, conn)
			}
		}
	}
}

func parseRESPArray(reader *bufio.Reader, line string) []string {
	numArgsStr := strings.TrimPrefix(line, "*")
	numArgs, err := strconv.Atoi(numArgsStr)
	if err != nil {
		return nil
	}

	var parts []string
	for i := 0; i < numArgs; i++ {
		// Skip bulk string length line
		_, err := reader.ReadString('\n')
		if err != nil {
			return nil
		}

		// Read actual content
		content, err := reader.ReadString('\n')
		if err != nil {
			return nil
		}
		parts = append(parts, strings.TrimSpace(content))
	}

	return parts
}
