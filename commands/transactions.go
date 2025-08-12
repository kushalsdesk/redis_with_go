package commands

import (
	"fmt"
	"net"
	"strconv"
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
		command != "MULTI"
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

		if parsedVal == 9223372036854775807 { // max int
			conn.Write([]byte("-ERR increment or decrement would overflow\r\n"))
			return
		}
		newValue = parsedVal + 1
	}
	store.Set(key, strconv.FormatInt(newValue, 10), 0)

	resp := fmt.Sprintf(":%d\r\n", newValue)
	conn.Write([]byte(resp))
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
