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
