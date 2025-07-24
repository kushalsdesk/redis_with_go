# Redis Phase 1: Technical Analysis & Phase 2 Preparation

## Phase 1 Implementation Review

### Architecture Overview
Your Redis implementation follows a clean, modular architecture:
- **main.go**: Entry point, minimal and focused
- **server.go**: Network layer handling TCP connections
- **handler.go**: Protocol parsing (RESP - Redis Serialization Protocol)
- **commands/**: Command implementations with dispatch pattern
- **store/**: In-memory data storage with TTL support

### Strong Points ✅

#### 1. **Proper Concurrency Handling**
```go
go handler.HandleConnection(conn)  // Each connection in separate goroutine
```
- Uses goroutines for concurrent client handling
- Thread-safe storage with `sync.RWMutex`
- Proper read/write locking in store operations

#### 2. **RESP Protocol Implementation**
Your RESP parsing correctly handles:
- Array format: `*2\r\n$4\r\nPING\r\n`
- Bulk strings with length prefixes
- Both inline and RESP array commands

#### 3. **TTL Implementation**
- Lazy expiration on GET operations
- Clean expired key removal
- Proper time handling with `time.Duration`

#### 4. **Error Handling**
- Consistent RESP error responses
- Input validation for command arguments
- Network error handling with connection cleanup

### What You Need to Change for Phase 2 🔧

Your current implementation works great for strings. For lists, you just need to:
1. **Modify store to handle different data types** (string vs list)
2. **Add type checking** in commands to prevent SET on a list key
3. **That's it for now** - keep iterating as you build!

## Preparing for Phase 2: Redis Lists

### Data Structure Design

#### Current Store Interface
```go
func Set(key, val string, ttl time.Duration)
func Get(key string) (string, bool)
```

#### Enhanced Store for Lists
```go
type Value struct {
    Type    ValueType  // STRING, LIST, etc.
    Data    interface{}
    Expiry  *time.Time
}

type List struct {
    Elements []string
    mutex    sync.RWMutex
}
```

### Key Phase 2 Challenges

#### 1. **List Operations Complexity**
- **LPUSH/RPUSH**: O(1) operations - Use slice append
- **LINDEX**: O(1) for positive, need length calculation for negative
- **LRANGE**: O(N) slice operations
- **LREM**: O(N) with element shifting

#### 2. **Blocking Operations**
```go
// BLPOP implementation concept
func handleBLPOP(args []string, conn net.Conn) {
    // Wait for list to have elements or timeout
    // Requires channels and goroutine coordination
}
```

#### 3. **Index Handling**
```go
func normalizeIndex(index, length int) int {
    if index < 0 {
        return length + index  // -1 becomes last element
    }
    return index
}
```

### Recommended Refactoring for Phase 2

#### 1. **Type System Enhancement**
```go
package store

type ValueType int
const (
    STRING ValueType = iota
    LIST
)

type RedisValue struct {
    Type   ValueType
    String string      // for STRING type
    List   []string    // for LIST type
    Expiry *time.Time
}
```

#### 2. **Command Interface**
```go
type CommandHandler interface {
    Execute(args []string, conn net.Conn) error
    ValidateArgs(args []string) error
}
```

#### 3. **Error Constants**
```go
const (
    ERR_WRONG_TYPE = "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n"
    ERR_SYNTAX     = "-ERR syntax error\r\n"
    ERR_ARGS       = "-ERR wrong number of arguments for '%s' command\r\n"
)
```

## Phase 2 Implementation Strategy

### Stage-by-Stage Approach

#### Stages 1-3: Basic List Creation
1. **LPUSH/RPUSH**: Modify store to handle list type
2. **Type checking**: Ensure operations match key types
3. **Multiple elements**: Handle variadic arguments

#### Stages 4-5: List Access
1. **LINDEX**: Implement positive/negative indexing
2. **LRANGE**: Slice operations with bounds checking

#### Stages 6-9: List Modification
1. **LPOP/RPOP**: Remove and return elements
2. **LLEN**: Return list length
3. **LREM**: Remove elements by value

#### Stages 10-11: Blocking Operations
1. **BLPOP/BRPOP**: Most complex - requires:
   - Client waiting queues
   - Timeout handling
   - Cross-goroutine communication

### Critical Implementation Notes

#### Memory Safety
```go
// Always check bounds
if index >= 0 && index < len(list.Elements) {
    return list.Elements[index]
}
```

#### Atomic Operations
```go
func (s *Store) ListPush(key string, elements []string, left bool) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    // Atomic list modification
}
```

#### Blocking Command Architecture
```go
type BlockingClient struct {
    Conn    net.Conn
    Keys    []string
    Timeout time.Duration
    Done    chan string
}

var blockingClients = make(map[string][]*BlockingClient)
```

## Next Steps Recommendation

1. **Refactor store package** to support multiple data types
2. **Implement basic list commands** (LPUSH, RPUSH, LLEN)
3. **Add comprehensive testing** for edge cases
4. **Tackle blocking operations** last (most complex)

Your foundation is solid - the modular design will serve you well as complexity increases. The concurrency patterns you've established will be crucial for blocking operations in phase 2.
