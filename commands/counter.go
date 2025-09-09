package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handleIncr(args []string, conn net.Conn) {
	if len(args) != 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'incr' command\r\n"))
		return
	}

	key := args[1]
	newValue, err := store.Increment(key)
	if err != nil {
		if strings.Contains(err.Error(), "WRONGTYPE") {
			conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
		} else if strings.Contains(err.Error(), "overflow") {
			conn.Write([]byte("-ERR increment or decrement would overflow\r\n"))
		} else {
			conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
		}
		return
	}

	resp := fmt.Sprintf(":%d\r\n", newValue)
	conn.Write([]byte(resp))
}

func handleDecr(args []string, conn net.Conn) {
	if len(args) != 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'decr' command\r\n"))
		return
	}

	key := args[1]
	newValue, err := store.Decrement(key)
	if err != nil {
		if strings.Contains(err.Error(), "WRONGTYPE") {
			conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
		} else if strings.Contains(err.Error(), "overflow") {
			conn.Write([]byte("-ERR increment or decrement would overflow\r\n"))
		} else {
			conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
		}
		return
	}

	resp := fmt.Sprintf(":%d\r\n", newValue)
	conn.Write([]byte(resp))
}

func handleIncrBy(args []string, conn net.Conn) {
	if len(args) == 2 {
		handleIncr(args, conn)
		return
	}

	if len(args) != 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'incrby' command\r\n"))
		return
	}

	key := args[1]
	amount, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
		return
	}

	if amount == 0 {
		currentVal, exists := store.Get(key)
		if !exists {
			conn.Write([]byte(":0\r\n"))
		} else {
			parsedVal, err := strconv.ParseInt(currentVal, 10, 64)
			if err != nil {
				conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
				return
			}
			resp := fmt.Sprintf(":%d\r\n", parsedVal)
			conn.Write([]byte(resp))
		}
		return
	}

	if amount < 0 {
		conn.Write([]byte("-ERR increment amount must be positive\r\n"))
		return
	}

	newValue, err := store.IncrementBy(key, amount)
	if err != nil {
		if strings.Contains(err.Error(), "WRONGTYPE") {
			conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
		} else if strings.Contains(err.Error(), "overflow") {
			conn.Write([]byte("-ERR increment or decrement would overflow\r\n"))
		} else {
			conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
		}
		return
	}

	resp := fmt.Sprintf(":%d\r\n", newValue)
	conn.Write([]byte(resp))
}

func handleDecrBy(args []string, conn net.Conn) {
	if len(args) == 2 {
		handleDecr(args, conn)
		return
	}

	if len(args) != 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'decrby' command\r\n"))
		return
	}

	key := args[1]
	amount, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
		return
	}

	newValue, err := store.DecrementBy(key, amount)
	if err != nil {
		if strings.Contains(err.Error(), "WRONGTYPE") {
			conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
		} else if strings.Contains(err.Error(), "overflow") {
			conn.Write([]byte("-ERR increment or decrement would overflow\r\n"))
		} else {
			conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
		}
		return
	}

	resp := fmt.Sprintf(":%d\r\n", newValue)
	conn.Write([]byte(resp))
}
