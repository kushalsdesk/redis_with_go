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
		// Clean up replica connection on disconnect
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
			// Parse RESP Array
			numArgsStr := strings.TrimPrefix(line, "*")
			numArgs, err := strconv.Atoi(numArgsStr)
			if err != nil {
				conn.Write([]byte("-ERR invalid array length\r\n"))
				return
			}

			var parts []string
			for i := 0; i < numArgs; i++ {
				// Read bulk string header: $<length>
				_, err := reader.ReadString('\n') // skip "$<len>"
				if err != nil {
					return
				}
				content, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				parts = append(parts, strings.TrimSpace(content))
			}

			// Check if we should queue this command instead of executing it
			if len(parts) > 0 && commands.ShouldQueueCommand(conn, strings.ToUpper(parts[0])) {
				commands.QueueCommand(conn, parts)
			} else {
				commands.Dispatch(parts, conn)
			}

		} else {
			// Inline command support
			parts := strings.Fields(line)
			if len(parts) > 0 {
				// Check if we should queue this command instead of executing it
				if commands.ShouldQueueCommand(conn, strings.ToUpper(parts[0])) {
					commands.QueueCommand(conn, parts)
				} else {
					commands.Dispatch(parts, conn)
				}
			}
		}
	}
}
