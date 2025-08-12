package commands

import (
	"fmt"
	"net"
	"strconv"

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
	resp := fmt.Sprintf(":%d\r\n", length)
	conn.Write([]byte(resp))

}

func handleRPush(args []string, conn net.Conn) {
	if len(args) < 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'rpush' command\r\n"))
		return
	}

	key := args[1]
	elements := args[2:]

	length := store.ListPush(key, elements, false)

	if length == -1 {
		conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
		return
	}

	resp := fmt.Sprintf(":%d\r\n", length)
	conn.Write([]byte(resp))
}

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

func handleLIndex(args []string, conn net.Conn) {
	if len(args) != 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'lindex' command\r\n"))
		return
	}

	key := args[1]
	indexStr := args[2]

	index, err := strconv.Atoi(indexStr)
	if err != nil {
		conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
		return
	}

	element, exists := store.ListIndex(key, index)

	if !exists {
		if store.GetListLength(key) == -1 {
			conn.Write([]byte("-WRONGTYPE operation against a key holding wrong type\r\n"))
			return
		}

		conn.Write([]byte("$-1\r\n"))
		return
	}

	resp := fmt.Sprintf("$%d\r\n%s\r\n", len(element), element)
	conn.Write([]byte(resp))

}

func handleLLen(args []string, conn net.Conn) {
	if len(args) != 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'llen' command\r\n"))
		return
	}

	key := args[1]
	length := store.GetListLength(key)

	if length == -1 {
		conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
		return
	}

	// Return length as integer
	resp := fmt.Sprintf(":%d\r\n", length)
	conn.Write([]byte(resp))
}

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

func handleLPop(args []string, conn net.Conn) {

	if len(args) < 2 || len(args) > 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'lpop' command\r\n"))
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
		element, exists := store.ListPop(key, true)
		if !exists {
			if store.GetListLength(key) == -1 {
				conn.Write([]byte("-WRONGTYPE operation against a key holding the wrong kind of value\r\n"))
				return
			}
			conn.Write([]byte("$-1\r\n"))
			return
		}
		resp := fmt.Sprintf("$%d\r\n%s\r\n", len(element), element)
		conn.Write([]byte(resp))
		return
	}

	//handling multiple element case
	elements, exists := store.ListPopMultiple(key, count, true)
	if !exists {
		if store.GetListLength(key) == -1 {
			conn.Write([]byte("-WRONGTYPE operation against a key holding the wrong kind of value\r\n"))
			return
		}
		conn.Write([]byte("*0\r\n"))
		return
	}

	// return a RESP array
	resp := fmt.Sprintf("*%d\r\n", len(elements))
	for _, element := range elements {
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(element), element)
	}
	conn.Write([]byte(resp))
}
