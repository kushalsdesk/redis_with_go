package commands

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/kushalsdesk/redis_with_go/store"
)

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
