package commands

import (
	"fmt"
	"net"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handleGet(args []string, conn net.Conn) {
	if len(args) < 2 {
		conn.Write([]byte("-ERR wrong number of arguments for GET\r\n"))
		return
	}
	key := args[1]
	val, ok := store.Get(key)
	if !ok {
		conn.Write([]byte("$-1\r\n"))
	} else {
		resp := fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
		conn.Write([]byte(resp))
	}
}
