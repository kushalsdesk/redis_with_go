package commands

import (
	"fmt"
	"net"
	"strconv"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handleLIndex(args []string, conn net.Conn) {
	if len(args) != 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'lindex' command\r\n"))
		return
	}

	key := args[1]
	indexStr := args[2]

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
		return
	}

	element, exists := store.ListIndex(key, index)

	if !exists {
		if store.GetListLength(key) == -1 {
			conn.Write([]byte("-WRONGTYPE operation against a key holding wrong type\r\n"))
			return
		}

		conn.Write([]byte("$-1\r\n"))
		return
	}

	resp := fmt.Sprintf("$%d\r\n%s\r\n", len(element), element)
	conn.Write([]byte(resp))

}
