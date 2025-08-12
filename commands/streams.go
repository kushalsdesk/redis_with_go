package commands

import (
	"fmt"
	"net"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handleXAdd(args []string, conn net.Conn) {
	// eg: XADD KEY id field1 field2 field3 field4
	if len(args) < 4 {
		conn.Write([]byte("-ERR wrong number of arguments for 'xadd' command\r\n"))
		return
	}

	key := args[1]
	id := args[2]
	fieldArgs := args[3:]

	// fields must come in pairs
	if len(fieldArgs)%2 != 0 {
		conn.Write([]byte("-ERR wrong number of arguments for 'xadd' command\r\n"))
		return
	}

	// parsing field-value pair
	fields := make(map[string]string)
	for i := 0; i < len(fieldArgs); i += 2 {
		fields[fieldArgs[i]] = fieldArgs[i+1]
	}

	resultID, err := store.StreamAdd(key, id, fields)
	if err != nil {
		if err.Error() == "WRONGTYPE Operation against a key holding the wrong kind of value" {
			conn.Write([]byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"))
		} else {
			conn.Write([]byte(fmt.Sprintf("-ERR %s\r\n", err.Error())))
		}
		return
	}
	resp := fmt.Sprintf("$%d\r\n%s\r\n", len(resultID), resultID)
	conn.Write([]byte(resp))

}

func handleXRange(args []string, conn net.Conn) {
	if len(args) < 4 {
		conn.Write([]byte("-ERR wrong number of arguments for 'xrange' command\r\n"))
		return
	}

	key := args[1]
	start := args[2]
	end := args[3]

	entries, err := store.StreamRange(key, start, end)
	if err != nil {
		if err.Error() == "WRONGTYPE Operation against a key holding the wrong kind of value" {
			conn.Write([]byte("WRONGTYPE Operation against a key holding the wrong kind of value"))
		} else {
			conn.Write([]byte(fmt.Sprintf("-ERR %s\r\n", err.Error())))
			// fmt.Fprintf(conn, "-ERR %s\r\n", err.Error())
		}
		return
	}

	resp := formatXRangeResponse(entries)
	conn.Write([]byte(resp))

}

func formatXRangeResponse(entries []store.StreamEntry) string {
	if len(entries) == 0 {
		return "*0\r\n"
	}

	response := fmt.Sprintf("*%d\r\n", len(entries))

	for _, entry := range entries {
		// Each entry is an array of [ID, [field1, value1, field2, value2, ...]]
		fieldCount := len(entry.Fields) * 2
		entryResponse := fmt.Sprintf("*2\r\n$%d\r\n%s\r\n*%d\r\n",
			len(entry.ID), entry.ID, fieldCount)

		for field, value := range entry.Fields {
			entryResponse += fmt.Sprintf("$%d\r\n%s\r\n$%d\r\n%s\r\n",
				len(field), field, len(value), value)
		}

		response += entryResponse
	}

	return response
}
