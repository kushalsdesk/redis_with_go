package commands

import (
	"fmt"
	"net"
	"strconv"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handleRPop(args []string, conn net.Conn) {
	if len(args) < 2 || len(args) > 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'rpop' commands\r\n"))
		return
	}

	key := args[1]
	count := 1

	if len(args) == 3 {
		var err error
		count, err = strconv.Atoi(args[2])
		if err != nil || count < 0 {
			conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
			return
		}
	}

	if count == 1 {
		element, exists := store.ListPop(key, false)
		if !exists {
			if store.GetListLength(key) == -1 {
				conn.Write([]byte("-WRONGTYPE operation against a key holding wrong kind of valu\r\n"))
				return
			}
			conn.Write([]byte("$-1\r\n"))
			return
		}
		resp := fmt.Sprintf("$%d\r\n%s\r\n", len(element), element)
		conn.Write([]byte(resp))
		return
	}

	elements, exists := store.ListPopMultiple(key, count, false)
	if !exists {
		if store.GetListLength(key) == -1 {
			conn.Write([]byte("-WRONGTYPE operation against a key holding wrong kind of valu\r\n"))
			return
		}
		conn.Write([]byte("*0\r\n"))
		return
	}

	resp := fmt.Sprintf("*%d\r\n", len(elements))
	for _, element := range elements {
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(element), element)
	}
	conn.Write([]byte(resp))
}
