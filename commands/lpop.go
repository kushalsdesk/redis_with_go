package commands

import (
	"fmt"
	"net"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handleLPop(args []string, conn net.Conn) {

	if len(args) != 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'lpop' command\r\n"))
		return
	}

	key := args[1]
	element, exists := store.ListPop(key, true)

	if !exists {
		if store.GetListLength(key) == -1 {
			conn.Write([]byte("-WRONGTYPE operation against a key holding the wrong kind of value\r\n"))
			return
		}
		conn.Write([]byte("$-1\r\n"))
		return
	}

	resp := fmt.Sprintf("$%d\r\n%s\r\n", len(element), element)
	conn.Write([]byte(resp))
}
