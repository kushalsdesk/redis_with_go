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

	// Execute the command first
	switch strings.ToUpper(args[0]) {
	case "PING":
		handlePing(conn)
	case "ECHO":
		handleEcho(args, conn)
	case "INFO":
		handleInfo(args, conn)
	case "SET":
		handleSet(args, conn)
		// PROPAGATE write commands after successful execution
		PropagateCommand(args)
	case "GET":
		handleGet(args, conn)
	case "LPUSH":
		handleLPush(args, conn)
		PropagateCommand(args) // PROPAGATE
	case "RPUSH":
		handleRPush(args, conn)
		PropagateCommand(args) // PROPAGATE
	case "LINDEX":
		handleLIndex(args, conn)
	case "LRANGE":
		handleLRange(args, conn)
	case "LLEN":
		handleLLen(args, conn)
	case "LPOP":
		handleLPop(args, conn)
		PropagateCommand(args) // PROPAGATE
	case "RPOP":
		handleRPop(args, conn)
		PropagateCommand(args) // PROPAGATE
	case "BLPOP":
		handleBLPop(args, conn)
	case "BRPOP":
		handleBRPop(args, conn)
	case "TYPE":
		handleType(args, conn)
	case "XADD":
		handleXAdd(args, conn)
		PropagateCommand(args) // PROPAGATE
	case "XRANGE":
		handleXRange(args, conn)
	case "XREAD":
		handleXRead(args, conn)
	case "INCR":
		handleIncr(args, conn)
		PropagateCommand(args) // PROPAGATE
	case "INCRBY":
		handleIncrBy(args, conn)
		PropagateCommand(args) // PROPAGATE
	case "DECR":
		handleDecr(args, conn)
		PropagateCommand(args) // PROPAGATE
	case "DECRBY":
		handleDecrBy(args, conn)
		PropagateCommand(args) // PROPAGATE
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
