package commands

import (
	"fmt"
	"net"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handleLPush(args []string, conn net.Conn) {
	if len(args) < 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'lpush' commmand\r\n"))
		return
	}

	key := args[1]
	elements := args[2:]

	length := store.ListPush(key, elements, true)

	if length == -1 {
		conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
		return
	}

	//return a new length as integer
	resp := []byte(fmt.Sprintf(":%d\r\n", length))
	conn.Write([]byte(resp))

}
