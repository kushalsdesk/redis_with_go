package commands

import (
	"fmt"
	"net"
)

func handleEcho(args []string, conn net.Conn) {
	if len(args) < 2 {
		conn.Write([]byte("-ERR wrong number of arguments\r\n"))
		return
	}
	msg := args[1]
	resp := fmt.Sprintf("$%d\r\n%s\r\n", len(msg), msg)
	conn.Write([]byte(resp))

}
