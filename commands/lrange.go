package commands

import (
	"fmt"
	"net"
	"strconv"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handleLRange(args []string, conn net.Conn) {
	if len(args) != 4 {
		conn.Write([]byte("-ERR wrong number of arguments for 'lrange' command\r\n"))
		return
	}

	key := args[1]
	startStr := args[2]
	stopStr := args[3]

	start, err1 := strconv.Atoi(startStr)
	stop, err2 := strconv.Atoi(stopStr)

	if err1 != nil || err2 != nil {
		conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
		return
	}

	elements, ok := store.ListRange(key, start, stop)

	if !ok {
		conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
		return
	}

	resp := fmt.Sprintf("*%d\r\n", len(elements))
	for _, element := range elements {
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(element), element)
	}

	conn.Write([]byte(resp))
}
