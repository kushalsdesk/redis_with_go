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
				conn.Write([]byte("-ERR syntax error\r\n"))
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
		handleBlockingXRead(streamKeys, streamIDs, blockMillis, count, conn)
		return
	}

	handleNonBlockingXRead(streamKeys, streamIDs, count, conn)

}

func handleNonBlockingXRead(streamKeys, streamIDs []string, count int, conn net.Conn) {
	results := make([]StreamReadResult, 0)

	for i, key := range streamKeys {
		startID := streamIDs[i]

		if startID == "$" {
			continue
		}
		entries, err := store.StreamReadFrom(key, startID, count)
		if err != nil {
			fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
			return
		}

		if len(entries) > 0 {
			results = append(results, StreamReadResult{
				StreamKey: key,
				Entries:   entries,
			})
		}
	}
	resp := formatXReadResponse(results)
	conn.Write([]byte(resp))
}

func handleBlockingXRead(streamKeys, streamIDs []string, blockMillis int64, count int, conn net.Conn) {
	//basic blocking mechanism
	timeout := time.Duration(blockMillis) * time.Millisecond
	if blockMillis == 0 {
		timeout = 0
	}
	startTime := time.Now()

	for {
		results := make([]StreamReadResult, 0)

		for i, key := range streamKeys {
			startID := streamIDs[i]

			if startID == "$" {
				lastID := store.GetStreamLastID(key)
				if lastID != "" {
					startID = lastID
				}
			}

			entries, err := store.StreamReadFrom(key, startID, count)
			if err != nil {
				fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
				return
			}
			if len(entries) > 0 {
				results = append(results, StreamReadResult{
					StreamKey: key,
					Entries:   entries,
				})
			}
		}

		if len(results) > 0 {
			response := formatXReadResponse(results)
			conn.Write([]byte(response))
			return
		}
		if timeout > 0 && time.Since(startTime) >= timeout {
			conn.Write([]byte("*-1\r\n"))
			return
		}

		time.Sleep(50 * time.Millisecond)
	}
}

func formatXReadResponse(results []StreamReadResult) string {
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
