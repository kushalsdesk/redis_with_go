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
	case "INFO":
		handleInfo(args, conn)
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
	case "BLPOP":
		handleBLPop(args, conn)
	case "BRPOP":
		handleBRPop(args, conn)
	case "TYPE":
		handleType(args, conn)
	case "XADD":
		handleXAdd(args, conn)
	case "XRANGE":
		handleXRange(args, conn)
	case "XREAD":
		handleXRead(args, conn)
	case "INCR":
		handleIncr(args, conn)
	case "INCRBY":
		handleIncrBy(args, conn)
	case "DECR":
		handleDecr(args, conn)
	case "DECRBY":
		handleDecrBy(args, conn)
	case "MULTI":
		handleMulti(args, conn)
	case "EXEC":
		handleExec(args, conn)
	case "DISCARD":
		handleDiscard(args, conn)
	case "UNDO":
		handleUndo(args, conn)
	case "PSYNC":
		handlePsync(args, conn)
	case "REPLCONF":
		handleReplconf(args, conn)
	default:
		conn.Write([]byte("-ERR unknown command\r\n"))
	}

}
