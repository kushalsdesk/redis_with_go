package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/kushalsdesk/redis_with_go/store"
)

type StreamReadResult struct {
	StreamKey string
	Entries   []store.StreamEntry
}

func handleXRead(args []string, conn net.Conn) {
	if len(args) < 4 {
		conn.Write([]byte("-ERR wrong number of arguments for 'xread' command\r\n"))
		return
	}

	var count int = -1         // no limit
	var blockMillis int64 = -1 // non-blocking
	var streamsIndex int = -1

	i := 1
	for i < len(args) {
		switch strings.ToUpper(args[i]) {
		case "COUNT":
			if i+1 >= len(args) {
				// conn.Write([]byte("-ERR syntax error\r\n"))
				fmt.Fprintf(conn, "-ERR syntax error\r\n")
				return
			}
			var err error
			count, err = strconv.Atoi(args[i+1])
			if err != nil || count <= 0 {
				conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
				return
			}
			i += 2
		case "BLOCK":
			if i+1 >= len(args) {
				conn.Write([]byte("-ERR syntax error\r\n"))
				return
			}
			var err error
			blockMillis, err = strconv.ParseInt(args[i+1], 10, 64)
			if err != nil || blockMillis < 0 {
				conn.Write([]byte("-ERR timeout is not a float or out range\r\n"))
				return
			}
			i += 2
		case "STREAMS":
			streamsIndex = i + 1
			i = len(args)
		default:
			fmt.Fprintf(conn, "-ERR unknown option: %s\r\n", args[i])
			return
		}
	}
	if streamsIndex == -1 {
		conn.Write([]byte("-ERR syntax error\r\n"))
		return
	}

	streamArgs := args[streamsIndex:]
	if len(streamArgs)%2 != 0 {
		conn.Write([]byte("-ERR Unbalanced XREAD list of streams: for each stream key an ID or '$' must be specified\r\n"))
		return
	}

	numStreams := len(streamArgs) / 2
	streamKeys := streamArgs[:numStreams]
	streamIDs := streamArgs[numStreams:]

	//basic blocking
	if blockMillis >= 0 {
		handleAdvancedBlockingXRead(streamKeys, streamIDs, blockMillis, count, conn)
		return
	}

	handleNonBlockingXRead(streamKeys, streamIDs, count, conn)

}

func handleNonBlockingXRead(streamKeys, streamIDs []string, count int, conn net.Conn) {
	results, _ := store.StreamReadFromImmediate(streamKeys, streamIDs, count)
	response := formatXReadResponse(results)
	fmt.Fprint(conn, response)

}

func handleAdvancedBlockingXRead(streamKeys, streamIDs []string, blockMillis int64, count int, conn net.Conn) {
	// trying immediate read
	results, hasData := store.StreamReadFromImmediate(streamKeys, streamIDs, count)
	if hasData {
		response := formatXReadResponse(results)
		fmt.Fprint(conn, response)
		return
	}

	var timeout time.Duration
	if blockMillis == 0 {
		timeout = 0
	} else {
		timeout = time.Duration(blockMillis) * time.Millisecond
	}

	client := store.RegisterStreamBlockingClient(streamKeys, streamIDs, count, timeout)
	defer store.UnregisterStreamBlockingClient(client)

	// Waiting for data/timeout
	if timeout > 0 {
		select {
		case result := <-client.Response:
			if result.Success {
				response := formatXReadResponse(result.Results)
				fmt.Fprint(conn, response)

			} else {
				fmt.Fprintf(conn, "*-1\r\n")
			}
		case <-time.After(timeout):
			fmt.Fprintf(conn, "*-1\r\n")
		}
	} else {
		// blocking indefinitely
		result := <-client.Response
		if result.Success {
			response := formatXReadResponse(result.Results)
			fmt.Fprint(conn, response)
		} else {
			fmt.Fprint(conn, "*-1\r\n")
		}
	}
}

func formatXReadResponse(results []store.StreamReadResult) string {
	if len(results) == 0 {
		return "*-1\r\n"
	}

	response := fmt.Sprintf("*%d\r\n", len(results))

	for _, result := range results {
		response += fmt.Sprintf("*2\r\n$%d\r\n%s\r\n*%d\r\n",
			len(result.StreamKey), result.StreamKey, len(result.Entries))

		for _, entry := range result.Entries {
			fieldCount := len(entry.Fields) * 2
			response += fmt.Sprintf("*2\r\n$%d\r\n%s\r\n*%d\r\n",
				len(entry.ID), entry.ID, fieldCount)

			for field, value := range entry.Fields {
				response += fmt.Sprintf("$%d\r\n%s\r\n$%d\r\n%s\r\n",
					len(field), field, len(value), value)
			}
		}
	}

	return response
}
