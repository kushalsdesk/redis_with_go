package commands

import (
	"fmt"
	"net"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handleType(args []string, conn net.Conn) {
	if len(args) != 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'type' command\r\n"))
	}

	key := args[1]
	keyType := store.GetKeyType(key)

	resp := fmt.Sprintf("+%s\r\n", keyType)
	conn.Write([]byte(resp))
}
