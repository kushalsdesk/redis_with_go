package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

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

func handleSet(args []string, conn net.Conn) {
	if len(args) < 3 {
		conn.Write([]byte("-ERR wrong number of arguments for SET\r\n"))
		return // ← this was missing
	}
	key := args[1]
	val := args[2]
	var ttl time.Duration

	if len(args) == 5 && strings.ToUpper(args[3]) == "EX" {
		seconds, err := strconv.Atoi(args[4])
		if err == nil {
			ttl = time.Duration(seconds) * time.Second
		}
	}

	store.Set(key, val, ttl)
	conn.Write([]byte("+OK\r\n"))
}
