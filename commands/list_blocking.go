package commands

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handleBLPop(args []string, conn net.Conn) {
	if len(args) < 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'blpop' command\r\n"))
		return
	}

	// last argument is timeout
	timeoutStr := args[len(args)-1]
	timeout, err := strconv.ParseFloat(timeoutStr, 64)
	if err != nil {
		conn.Write([]byte("-ERR timeout is not a float or out of range\r\n"))
		return
	}

	// as Keys are all arguments except the last one(timeout)
	keys := args[1 : len(args)-1]

	//tyring out immediate pop first
	key, element, found := store.ListBlockingPopImmediate(keys, true)
	if found {
		resp := fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(element), element)
		conn.Write([]byte(resp))
		return
	}

	// if timeout is 0, block indefinitely
	var timeoutDuration time.Duration
	if timeout == 0 {
		timeoutDuration = time.Duration(0)
	} else {
		timeoutDuration = time.Duration(timeout * float64(time.Second))
	}

	//Register for blocking
	client := store.RegisterBlockingClient(keys, true, timeoutDuration)
	defer store.UnregisterBlockingClient(client)

	//wait for result or timeout
	if timeout == 0 {
		result := <-client.Response
		if result.Success {
			resp := fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(result.Key), result.Key, len(result.Value), result.Value)
			conn.Write([]byte(resp))
		} else {
			conn.Write([]byte("$-1\r\n"))
		}
	} else {
		// block with timeout
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
