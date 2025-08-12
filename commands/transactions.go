package commands

import (
	"fmt"
	"net"
	"strconv"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handleIncr(args []string, conn net.Conn) {
	if len(args) != 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'incr' command\r\n"))
		return
	}

	key := args[1]

	currentVal, exists := store.Get(key)
	var newValue int64

	if !exists {
		newValue = 1
	} else {
		parsedVal, err := strconv.ParseInt(currentVal, 10, 64)
		if err != nil {
			conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
			return
		}

		if parsedVal == 9223372036854775807 { // max int
			conn.Write([]byte("-ERR increment or decrement would overflow\r\n"))
			return
		}
		newValue = parsedVal + 1
	}
	store.Set(key, strconv.FormatInt(newValue, 10), 0)

	resp := fmt.Sprintf(":%d\r\n", newValue)
	conn.Write([]byte(resp))
}
