# Redis Clone with Go

[![progress-banner](https://backend.codecrafters.io/progress/redis/0a412eea-657f-434d-b2cc-b7352c66c04f)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)

This is a starting point for Go solutions to the ["Build Your Own Redis" Challenge](https://codecrafters.io/challenges/redis).

## Overview

Redis is an in-memory data structure store often used as a database, cache, message broker and streaming engine. In this challenge you'll build your own Redis server that is capable of serving basic commands, reading RDB files and more.

Along the way, we'll learn about TCP servers, the Redis Protocol, data structures, concurrency patterns, and advanced Redis features like transactions and streams.

## Implementation Progress

> **Difficulty Levels:** 🟩 Easy | 🟨 Medium | 🟥 Hard

### [Phase 1: Basic Server & String Operations](./docs/phase1.md) - **✅ COMPLETED**

- [x] Bind to a port ................................................... 🟩⬜⬜
- [x] Respond to PING .................................................. 🟩⬜⬜
- [x] Respond to multiple PINGS ........................................ 🟩⬜⬜
- [x] Handle concurrent clients ........................................ 🟩🟨⬜
- [x] Implement the ECHO command ....................................... 🟩⬜⬜
- [x] Implement the SET & GET command .................................. 🟩🟨⬜
- [x] Expiry ........................................................... 🟩🟨⬜

### [Phase 2: Lists & Blocking Operations](./docs/phase2.md) - **✅ COMPLETED**

- [x] Create a list ................................................... 🟩⬜⬜
- [x] Append an element (RPUSH) ....................................... 🟩⬜⬜
- [x] Append multiple elements ........................................ 🟩⬜⬜
- [x] List elements (positive indexes) ................................ 🟩🟨⬜
- [x] List elements (negative indexes) ................................ 🟩🟨⬜
- [x] Prepend elements (LPUSH) ........................................ 🟩⬜⬜
- [x] Query list length ............................................... 🟩⬜⬜
- [x] Remove an element ............................................... 🟩🟨⬜
- [x] Remove multiple elements ........................................ 🟩🟨⬜
- [x] Blocking retrieval (BLPOP/BRPOP) ................................ 🟩🟨🟥
- [x] Blocking retrieval with timeout ................................. 🟩🟨🟥

### [Phase 3: Streams & Advanced Blocking](./docs/phase3.md) - **✅ COMPLETED**

- [x] The TYPE command ................................................ 🟩⬜⬜
- [x] Create a stream (XADD) .......................................... 🟩🟨⬜
- [x] Validating entry IDs ............................................ 🟩🟨🟥
- [x] Partially auto-generate IDs ..................................... 🟩🟨⬜
- [x] Fully auto-generate IDs ......................................... 🟩🟨⬜
- [x] Query entries into stream (XRANGE) .............................. 🟩🟨⬜
- [x] Query with - .................................................... 🟩🟨⬜
- [x] Query with + .................................................... 🟩🟨⬜
- [x] Query single stream using XREAD ................................. 🟩🟨🟥
- [x] Query multiple streams using XREAD .............................. 🟩🟨🟥
- [x] Blocking reads with timeout ..................................... 🟩🟨🟥
- [x] Blocking reads without timeout (BLOCK 0) ........................ 🟩🟨🟥
- [x] Blocking reads using $ .......................................... 🟩🟨🟥

### [Phase 4: Transactions](./docs/phase4.md) - **🚧 COMING NEXT**

- [ ] The INCR command (1/3) .......................................... 🟩⬜⬜
- [ ] The INCR command (2/3) .......................................... 🟩🟨⬜
- [ ] The INCR command (3/3) .......................................... 🟩🟨⬜
- [ ] The MULTI command ............................................... 🟩🟨⬜
- [ ] The EXEC command ................................................ 🟩🟨🟥
- [ ] Empty transaction ............................................... 🟩🟨⬜
- [ ] Queueing commands ............................................... 🟩🟨🟥
- [ ] Executing a transaction ......................................... 🟩🟨🟥
- [ ] The DISCARD command ............................................. 🟩🟨⬜
- [ ] Failures within transactions .................................... 🟩🟨🟥
- [ ] Multiple transactions ........................................... 🟩🟨🟥

## Project Structure

```
redis_with_go/
├── app/
│   └── main.go                    # Entry point
├── server/
│   ├── server.go                  # TCP server setup
│   └── handler/
│       └── handler.go             # RESP protocol parsing & connection handling
├── commands/
│   ├── dispatch.go                # Command routing & distribution
│   ├── basic.go                   # PING, ECHO commands
│   ├── strings.go                 # SET, GET commands
│   ├── lists.go                   # LPUSH, RPUSH, LRANGE, LLEN, etc.
│   ├── list_blocking.go           # BLPOP, BRPOP commands
│   ├── streams.go                 # XADD, XRANGE commands
│   ├── stream_blocking.go         # XREAD (blocking) commands
│   └── utils.go                   # TYPE command & utilities
├── store/                         # Refactored storage layer
│   ├── core.go                    # Core data structures & utilities
│   ├── string_ops.go              # String operations (SET, GET)
│   ├── list_ops.go                # List operations (PUSH, POP, RANGE)
│   ├── list_blocking.go           # List blocking operations (BLPOP, BRPOP)
│   ├── stream_ops.go              # Stream operations (XADD, XRANGE)
│   └── stream_blocking.go         # Stream blocking operations (XREAD)
├── docs/
│   ├── phase1.md                 # Bundle 1 implementation details
│   ├── phase2.md                 # Bundle 2 implementation details
│   └── phase3.md                 # Bundle 3 implementation details
└── README.md
```

## Getting Started

### 1. Setup & Installation

```bash
# Clone the repository
git clone https://github.com/kushalsdesk/redis_with_go
cd redis_with_go

# Run the server
go run app/main.go

# Server will start on port 6379
# Output: Server listening on :6379
```

### 2. Connect with Redis CLI

```bash
redis-cli -p 6379
```

## Comprehensive Test Cases

### **Phase 1: Basic Operations**

```bash
# Basic connectivity
127.0.0.1:6379> PING
PONG

# Echo command
127.0.0.1:6379> ECHO "Hello World"
"Hello World"

# String operations
127.0.0.1:6379> SET mykey "hello"
OK
127.0.0.1:6379> GET mykey
"hello"

# String with TTL (5 second expiry)
127.0.0.1:6379> SET tempkey "temporary" EX 5
OK
127.0.0.1:6379> GET tempkey
"temporary"
# Wait 5 seconds...
127.0.0.1:6379> GET tempkey
(nil)

# Type checking
127.0.0.1:6379> SET stringkey "value"
OK
127.0.0.1:6379> TYPE stringkey
string
127.0.0.1:6379> TYPE nonexistent
none
```

### **Phase 2: Lists & Blocking**

```bash
# List creation and basic operations
127.0.0.1:6379> LPUSH mylist "world"
(integer) 1
127.0.0.1:6379> LPUSH mylist "hello"
(integer) 2
127.0.0.1:6379> RPUSH mylist "!"
(integer) 3

# List querying
127.0.0.1:6379> LRANGE mylist 0 -1
1) "hello"
2) "world"
3) "!"
127.0.0.1:6379> LLEN mylist
(integer) 3
127.0.0.1:6379> LINDEX mylist 1
"world"

# List popping
127.0.0.1:6379> LPOP mylist
"hello"
127.0.0.1:6379> RPOP mylist
"!"

# Type checking for lists
127.0.0.1:6379> TYPE mylist
list

# Blocking operations (test with multiple terminals)
# Terminal 1:
127.0.0.1:6379> BLPOP waitlist 5
# (waits for up to 5 seconds)

# Terminal 2 (while Terminal 1 is waiting):
127.0.0.1:6379> LPUSH waitlist "data"
(integer) 1

# Terminal 1 immediately receives:
1) "waitlist"
2) "data"

# Infinite blocking (BLOCK 0)
# Terminal 1:
127.0.0.1:6379> BRPOP infinitelist 0
# (waits indefinitely)

# Terminal 2:
127.0.0.1:6379> RPUSH infinitelist "finally"
(integer) 1

# Terminal 1 immediately receives:
1) "infinitelist"
2) "finally"
```

###  **Phase 3: Streams & Advanced Blocking**

```bash
# Stream creation with auto-generated IDs
127.0.0.1:6379> XADD mystream * name "Alice" age "30"
"1754967302780-0"
127.0.0.1:6379> XADD mystream * name "Bob" age "25"
"1754967308123-0"

# Stream creation with custom IDs
127.0.0.1:6379> XADD teststream 1000-0 event "start"
"1000-0"
127.0.0.1:6379> XADD teststream 1000-1 event "progress"
"1000-1"

# Partial auto-generation (timestamp-sequence)
127.0.0.1:6379> XADD partialstream 2000-* action "create"
"2000-0"
127.0.0.1:6379> XADD partialstream 2000-* action "update"
"2000-1"

# Stream querying
127.0.0.1:6379> XRANGE mystream - +
1) 1) "1754967302780-0"
   2) 1) "name"
      2) "Alice"
      3) "age"
      4) "30"
2) 1) "1754967308123-0"
   2) 1) "name"
      2) "Bob"
      3) "age"
      4) "25"

# Stream type checking
127.0.0.1:6379> TYPE mystream
stream

# XREAD - reading from specific ID
127.0.0.1:6379> XREAD STREAMS mystream 1754967302780-0
1) 1) "mystream"
   2) 1) 1) "1754967308123-0"
         2) 1) "name"
            2) "Bob"
            3) "age"
            4) "25"

# XREAD with COUNT limit
127.0.0.1:6379> XREAD COUNT 1 STREAMS mystream 0-0
1) 1) "mystream"
   2) 1) 1) "1754967302780-0"
         2) 1) "name"
            2) "Alice"
            3) "age"
            4) "30"

#  **Advanced Blocking Features**

# Blocking XREAD with timeout (test with multiple terminals)
# Terminal 1:
127.0.0.1:6379> XREAD BLOCK 5000 STREAMS livestream $
# (waits for up to 5 seconds for new data)

# Terminal 2 (while Terminal 1 is waiting):
127.0.0.1:6379> XADD livestream * event "real-time" data "immediate"
"1754967414819-0"

# Terminal 1 immediately receives (within ~1 second):
1) 1) "livestream"
   2) 1) 1) "1754967414819-0"
         2) 1) "event"
            2) "real-time"
            3) "data"
            4) "immediate"

# Infinite blocking (BLOCK 0)
# Terminal 1:
127.0.0.1:6379> XREAD BLOCK 0 STREAMS waitstream $
# (waits indefinitely)

# Terminal 2:
127.0.0.1:6379> XADD waitstream * message "finally here"
"1754967420156-0"

# Terminal 1 immediately receives:
1) 1) "waitstream"
   2) 1) 1) "1754967420156-0"
         2) 1) "message"
            2) "finally here"

# Multiple streams blocking
# Terminal 1:
127.0.0.1:6379> XREAD BLOCK 10000 STREAMS stream1 stream2 $ $

# Terminal 2:
127.0.0.1:6379> XADD stream2 * data "from stream2"

# Terminal 1 immediately receives data from stream2
```
## Performance Features

## 🚀 **Real-time Notifications**
- **Immediate response**: Blocking commands return instantly when data arrives
- **No polling**: Event-driven architecture using Go channels
- **Concurrent safety**: Thread-safe operations with proper mutex usage

### 💪 **Concurrent Client Support**
- **Multi-client**: Multiple redis-cli connections work simultaneously
- **Independent blocking**: Each client can block on different keys/streams
- **Resource cleanup**: Automatic client cleanup on disconnection

### ⚡ **Memory Efficiency**
- **Modular storage**: Separated concerns across focused modules
- **TTL support**: Automatic expiry and cleanup
- **Type safety**: Strong typing for different Redis data structures

## Key Learning Outcomes

- **TCP Server Programming**: Building concurrent network servers in Go
- **Protocol Implementation**: Redis RESP (Redis Serialization Protocol)
- **Data Structures**: Implementing Redis data types (strings, lists, streams)
- **Concurrency**: Goroutines, channels, mutexes, and thread-safe operations
- **Memory Management**: TTL implementation and expiration handling
- **Advanced Patterns**: Real-time blocking operations, event-driven architecture
- **Software Architecture**: Clean code principles, modular design, separation of concerns
