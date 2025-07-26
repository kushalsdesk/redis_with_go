package commands

import (
	"net"
	"strings"
)

func Dispatch(args []string, conn net.Conn) {
	if len(args) == 0 {
		conn.Write([]byte("-ERR unknown command\r\n"))
		return
	}

	switch strings.ToUpper(args[0]) {
	case "PING":
		handlePing(conn)
	case "ECHO":
		handleEcho(args, conn)
	case "SET":
		handleSet(args, conn)
	case "GET":
		handleGet(args, conn)
	case "LPUSH":
		handleLPush(args, conn)
	case "RPUSH":
		handleRPush(args, conn)
	case "LINDEX":
		handleLIndex(args, conn)
	case "LRANGE":
		handleLRange(args, conn)
	case "LLEN":
		handleLLen(args, conn)
	case "LPOP":
		handleLPop(args, conn)
	case "RPOP":
		handleRPop(args, conn)

	default:
		conn.Write([]byte("-ERR unknown command\r\n"))
	}

}
