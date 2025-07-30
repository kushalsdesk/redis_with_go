package commands

import (
	"fmt"
	"net"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handleRPush(args []string, conn net.Conn) {
	if len(args) < 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'rpush' command\r\n"))
		return
	}

	key := args[1]
	elements := args[2:]

	length := store.ListPush(key, elements, false)

	if length == -1 {
		conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
		return
	}

	resp := fmt.Sprintf(":%d\r\n", length)
	conn.Write([]byte(resp))
}
