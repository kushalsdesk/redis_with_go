package commands

import (
	"fmt"
	"net"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handleLLen(args []string, conn net.Conn) {
	if len(args) != 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'llen' command\r\n"))
		return
	}

	key := args[1]
	length := store.GetListLength(key)

	if length == -1 {
		conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
		return
	}

	// Return length as integer
	resp := fmt.Sprintf(":%d\r\n", length)
	conn.Write([]byte(resp))
}
