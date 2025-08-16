package commands

import (
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kushalsdesk/redis_with_go/store"
)

type TransactionState struct {
	InTransaction  bool
	QueuedCommands [][]string
}

var (
	transactionStates = make(map[net.Conn]*TransactionState)
	transactionMutex  sync.RWMutex
)

func getTransactionState(conn net.Conn) *TransactionState {
	transactionMutex.RLock()
	defer transactionMutex.RUnlock()

	state, exists := transactionStates[conn]
	if !exists {
		return &TransactionState{
			InTransaction:  false,
			QueuedCommands: [][]string{},
		}
	}
	return state
}
func setTransactionState(conn net.Conn, state *TransactionState) {

	transactionMutex.Lock()
	defer transactionMutex.Unlock()
	transactionStates[conn] = state
}

func clearTransactionState(conn net.Conn) {

	transactionMutex.Lock()
	defer transactionMutex.Unlock()
	delete(transactionStates, conn)

}

func ShouldQueueCommand(conn net.Conn, command string) bool {

	state := getTransactionState(conn)
	return state.InTransaction &&
		command != "EXEC" &&
		command != "DISCARD" &&
		command != "MULTI" &&
		command != "UNDO"
}

func QueueCommand(conn net.Conn, args []string) {
	transactionMutex.Lock()
	defer transactionMutex.Unlock()

	state, exists := transactionStates[conn]
	if !exists {
		state = &TransactionState{
			InTransaction:  true,
			QueuedCommands: [][]string{},
		}
	}

	state.QueuedCommands = append(state.QueuedCommands, args)
	transactionStates[conn] = state

	conn.Write([]byte("+QUEUED\r\n"))

}

func handleMulti(args []string, conn net.Conn) {
	if len(args) != 1 {
		conn.Write([]byte("-ERR wrong number of arguments for 'multi' command\r\n"))
		return
	}

	state := getTransactionState(conn)

	if state.InTransaction {
		conn.Write([]byte("-ERR MULTI calls can not be nested\r\n"))
		return
	}

	newState := &TransactionState{
		InTransaction:  true,
		QueuedCommands: [][]string{},
	}
	setTransactionState(conn, newState)

	conn.Write([]byte("+OK\r\n"))
}

type MockConn struct {
	responses []string
}

// needed  interfaces to imitate -> conn net.Conn
func (m *MockConn) Write(b []byte) (int, error) {
	m.responses = append(m.responses, string(b))
	return len(b), nil

}
func (m *MockConn) Read(b []byte) (int, error)         { return 0, nil }
func (m *MockConn) Close() error                       { return nil }
func (m *MockConn) LocalAddr() net.Addr                { return nil }
func (m *MockConn) RemoteAddr() net.Addr               { return nil }
func (m *MockConn) SetDeadline(t time.Time) error      { return nil }
func (m *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConn) SetWriteDeadline(t time.Time) error { return nil }

func handleExec(args []string, conn net.Conn) {
	if len(args) != 1 {
		conn.Write([]byte("-ERR wrong number of arguments for 'exec' command\r\n"))
		return
	}

	state := getTransactionState(conn)

	if !state.InTransaction {
		conn.Write([]byte("-ERR EXEC without MULTI\r\n"))
		return
	}

	if len(state.QueuedCommands) == 0 {
		clearTransactionState(conn)
		conn.Write([]byte("*0\r\n"))
		return
	}

	results := make([]string, len(state.QueuedCommands))

	for i, queueArgs := range state.QueuedCommands {
		mockConn := &MockConn{responses: []string{}}

		Dispatch(queueArgs, mockConn)

		if len(mockConn.responses) > 0 {
			results[i] = mockConn.responses[0]
		} else {
			results[i] = "+OK\r\n"
		}
	}

	clearTransactionState(conn)

	resp := fmt.Sprintf("*%d\r\n", len(results))
	for _, result := range results {
		resp += result
	}

	conn.Write([]byte(resp))
}

func handleDiscard(args []string, conn net.Conn) {
	if len(args) != 1 {
		conn.Write([]byte("-ERR wrong number of arguments for 'discard' command\r\n"))
		return
	}

	state := getTransactionState(conn)

	if !state.InTransaction {
		conn.Write([]byte("-ERR DISCARD without MULTI\r\n"))
		return
	}

	clearTransactionState(conn)

	conn.Write([]byte("+OK\r\n"))
}

func handleUndo(args []string, conn net.Conn) {
	undoCount := 1
	if len(args) == 2 {
		count, err := strconv.Atoi(args[1])
		if err != nil || count <= 0 {
			conn.Write([]byte("-ERR invalid undo count\r\n"))
			return
		}
		undoCount = count
	} else if len(args) > 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'undo' command\r\n"))
		return
	}

	state := getTransactionState(conn)
	if !state.InTransaction {
		conn.Write([]byte("-ERR UNDO without MULTI\r\n"))
		return
	}

	if len(state.QueuedCommands) == 0 {
		conn.Write([]byte("*0\r\n"))
		return
	}

	if undoCount > len(state.QueuedCommands) {
		resp := fmt.Sprintf("-ERR cannot undo %d commands, only %d queued\r\n", undoCount, len(state.QueuedCommands))
		conn.Write([]byte(resp))
		return
	}

	// Get the commands to be removed
	removedCommands := state.QueuedCommands[len(state.QueuedCommands)-undoCount:]

	// Remove the commands from queue
	transactionMutex.Lock()
	state.QueuedCommands = state.QueuedCommands[:len(state.QueuedCommands)-undoCount]
	transactionStates[conn] = state
	transactionMutex.Unlock()

	// Format: *N\r\n where N = number of elements in response

	totalElements := undoCount + 2
	resp := fmt.Sprintf("*%d\r\n", totalElements)

	// Add summary as bulk string
	summary := fmt.Sprintf("Removed %d commands:", undoCount)
	resp += fmt.Sprintf("$%d\r\n%s\r\n", len(summary), summary)

	// Add each removed command as bulk string
	for _, cmd := range removedCommands {
		cmdStr := strings.Join(cmd, " ")
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(cmdStr), cmdStr)
	}

	// Add remaining count info as bulk string
	remainingInfo := fmt.Sprintf("%d commands remaining in queue", len(state.QueuedCommands))
	resp += fmt.Sprintf("$%d\r\n%s\r\n", len(remainingInfo), remainingInfo)

	conn.Write([]byte(resp))
}

func handleIncr(args []string, conn net.Conn) {
	if len(args) != 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'incr' command\r\n"))
		return
	}

	key := args[1]

	currentVal, exists := store.Get(key)
	var newValue int64

	if !exists {
		newValue = 1
	} else {
		parsedVal, err := strconv.ParseInt(currentVal, 10, 64)
		if err != nil {
			conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
			return
		}

		if parsedVal == math.MaxInt64 {
			conn.Write([]byte("-ERR increment or decrement would overflow\r\n"))
			return
		}
		newValue = parsedVal + 1
	}
	store.Set(key, strconv.FormatInt(newValue, 10), 0)

	resp := fmt.Sprintf(":%d\r\n", newValue)
	conn.Write([]byte(resp))
}

// DECR command - decrement by 1
func handleDecr(args []string, conn net.Conn) {
	if len(args) != 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'decr' command\r\n"))
		return
	}

	key := args[1]
	currentVal, exists := store.Get(key)

	var newValue int64

	if !exists {
		newValue = -1 // Non-existent key becomes -1 (opposite of INCR)
	} else {
		parsedVal, err := strconv.ParseInt(currentVal, 10, 64)
		if err != nil {
			conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
			return
		}

		// Check for underflow
		if parsedVal == math.MinInt64 {
			conn.Write([]byte("-ERR increment or decrement would overflow\r\n"))
			return
		}

		newValue = parsedVal - 1
	}

	store.Set(key, strconv.FormatInt(newValue, 10), 0)

	resp := fmt.Sprintf(":%d\r\n", newValue)
	conn.Write([]byte(resp))
}

// INCRBY command - increment by specified amount
func handleIncrBy(args []string, conn net.Conn) {
	if len(args) == 2 {
		handleIncr(args, conn)
		return
	}

	if len(args) != 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'incrby' command\r\n"))
		return
	}

	key := args[1]
	amount, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
		return
	}

	if amount == 0 {
		conn.Write([]byte("$-1(Use 'INCR' command instead)\r\n"))
		return
	}

	if amount < 0 {
		conn.Write([]byte("-ERR increment amount must be positive\r\n"))
		return
	}

	currentVal, exists := store.Get(key)
	var newValue int64

	if !exists {
		newValue = amount
	} else {
		parsedVal, err := strconv.ParseInt(currentVal, 10, 64)
		if err != nil {
			conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
			return
		}

		// Check for overflow
		if parsedVal > 0 && amount > math.MaxInt64-parsedVal {
			conn.Write([]byte("-ERR increment or decrement would overflow\r\n"))
			return
		}

		newValue = parsedVal + amount
	}

	store.Set(key, strconv.FormatInt(newValue, 10), 0)

	resp := fmt.Sprintf(":%d\r\n", newValue)
	conn.Write([]byte(resp))
}

// DECRBY command - decrement by specified amount (supports positive and negative)
func handleDecrBy(args []string, conn net.Conn) {
	if len(args) == 2 {
		handleDecr(args, conn)
		return
	}

	if len(args) != 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'decrby' command\r\n"))
		return
	}

	key := args[1]
	amount, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
		return
	}

	currentVal, exists := store.Get(key)
	var newValue int64

	if !exists {
		newValue = -amount
	} else {
		parsedVal, err := strconv.ParseInt(currentVal, 10, 64)
		if err != nil {
			conn.Write([]byte("-ERR value is not an integer or out of range\r\n"))
			return
		}

		if amount > 0 && parsedVal < math.MinInt64+amount {
			conn.Write([]byte("-ERR increment or decrement would overflow\r\n"))
			return
		}
		if amount < 0 && parsedVal > 9223372036854775807+amount {
			conn.Write([]byte("-ERR increment or decrement would overflow\r\n"))
			return
		}

		newValue = parsedVal - amount
	}

	store.Set(key, strconv.FormatInt(newValue, 10), 0)

	resp := fmt.Sprintf(":%d\r\n", newValue)
	conn.Write([]byte(resp))
}
