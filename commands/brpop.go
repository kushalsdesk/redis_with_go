package commands

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handleBRPop(args []string, conn net.Conn) {
	if len(args) < 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'brpop' command\r\n"))
		return
	}

	timeoutStr := args[len(args)-1]
	timeout, err := strconv.ParseFloat(timeoutStr, 64)
	if err != nil {
		conn.Write([]byte("-ERR timeout is not a float or out of range\r\n"))
		return
	}

	keys := args[1 : len(args)-1]
	key, element, found := store.ListBlockingPopImmediate(keys, false)
	if found {
		resp := fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(element), element)
		conn.Write([]byte(resp))
		return
	}

	var timeoutDuration time.Duration
	if timeout == 0 {
		timeoutDuration = time.Duration(0)
	} else {
		timeoutDuration = time.Duration(timeout * float64(time.Second))
	}

	// Regsitering for blocking
	client := store.RegisterBlockingClient(keys, false, timeoutDuration)
	defer store.UnregisterBlockingClient(client)

	// waiting for result or timeout
	if timeout == 0 {
		result := <-client.Response
		if result.Success {
			resp := fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(result.Key), result.Key, len(result.Value), result.Value)
			conn.Write([]byte(resp))
		} else {
			conn.Write([]byte("$-1\r\n"))
		}
	} else {
		select {
		case result := <-client.Response:
			if result.Success {
				resp := fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(result.Key), result.Key, len(result.Value), result.Value)
				conn.Write([]byte(resp))
			} else {
				conn.Write([]byte("$-1\r\n"))
			}
		case <-time.After(timeoutDuration):
			conn.Write([]byte("$-1\r\n"))
		}
	}
}
