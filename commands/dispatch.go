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

	command := strings.ToUpper(args[0])

	switch command {
	// Read commands (no propagation)
	case "PING":
		handlePing(conn)
	case "ECHO":
		handleEcho(args, conn)
	case "INFO":
		handleInfo(args, conn)
	case "GET":
		handleGet(args, conn)
	case "LINDEX":
		handleLIndex(args, conn)
	case "LRANGE":
		handleLRange(args, conn)
	case "LLEN":
		handleLLen(args, conn)
	case "TYPE":
		handleType(args, conn)
	case "XRANGE":
		handleXRange(args, conn)
	case "XREAD":
		handleXRead(args, conn)
	case "BLPOP":
		handleBLPop(args, conn)
	case "BRPOP":
		handleBRPop(args, conn)

	// Write commands (with propagation)
	case "SET":
		handleSet(args, conn)
		PropagateCommand(args)
	case "LPUSH":
		handleLPush(args, conn)
		PropagateCommand(args)
	case "RPUSH":
		handleRPush(args, conn)
		PropagateCommand(args)
	case "LPOP":
		handleLPop(args, conn)
		PropagateCommand(args)
	case "RPOP":
		handleRPop(args, conn)
		PropagateCommand(args)
	case "XADD":
		handleXAdd(args, conn)
		PropagateCommand(args)
	case "INCR":
		handleIncr(args, conn)
		PropagateCommand(args)
	case "INCRBY":
		handleIncrBy(args, conn)
		PropagateCommand(args)
	case "DECR":
		handleDecr(args, conn)
		PropagateCommand(args)
	case "DECRBY":
		handleDecrBy(args, conn)
		PropagateCommand(args)

	// Replication commands
	case "PSYNC":
		handlePsync(args, conn)
	case "REPLCONF":
		handleReplconf(args, conn)

	default:
		conn.Write([]byte("-ERR unknown command\r\n"))
	}
}
