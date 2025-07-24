package handler

import (
	"bufio"
	"net"
	"strconv"
	"strings"

	"github.com/kushalsdesk/redis_with_go/commands"
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()
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

			commands.Dispatch(parts, conn)
		} else {
			// Inline command support
			parts := strings.Fields(line)
			if len(parts) > 0 {
				commands.Dispatch(parts, conn)
			}
		}
	}
}
