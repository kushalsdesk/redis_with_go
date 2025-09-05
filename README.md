# Redis Clone with Go

[![progress-banner](https://backend.codecrafters.io/progress/redis/0a412eea-657f-434d-b2cc-b7352c66c04f)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)

This is a starting point for Go solutions to the ["Build Your Own Redis" Challenge](https://codecrafters.io/challenges/redis).

## Overview

Redis is an in-memory data structure store often used as a database, cache, message broker and streaming engine. In this challenge you'll build your own Redis server that is capable of serving basic commands, reading RDB files and more.

Along the way, we'll learn about TCP servers, the Redis Protocol, data structures, concurrency patterns, and advanced Redis features like transactions and streams.

## Implementation Progress

> **Difficulty Levels:** 🟩 Easy | 🟨 Medium | 🟥 Hard

### [Phase 1: Basic Server & String Operations](./docs/phase1.md) - **✅ COMPLETED**

- [x] Bind to a port ................................................... 🟩
- [x] Respond to PING .................................................. 🟩
- [x] Respond to multiple PINGS ........................................ 🟩
- [x] Handle concurrent clients ........................................ 🟨
- [x] Implement the ECHO command ....................................... 🟩
- [x] Implement the SET & GET command .................................. 🟨
- [x] Expiry ........................................................... 🟨

### [Phase 2: Lists & Blocking Operations](./docs/phase2.md) - **✅ COMPLETED**

- [x] Create a list ................................................... 🟩
- [x] Append an element (RPUSH) ....................................... 🟩
- [x] Append multiple elements ........................................ 🟩
- [x] List elements (positive indexes) ................................ 🟩
- [x] List elements (negative indexes) ................................ 🟩
- [x] Prepend elements (LPUSH) ........................................ 🟩
- [x] Query list length ............................................... 🟩
- [x] Remove an element ............................................... 🟨
- [x] Remove multiple elements ........................................ 🟨
- [x] Blocking retrieval (BLPOP/BRPOP) ................................ 🟥
- [x] Blocking retrieval with timeout ................................. 🟥

### [Phase 3: Streams & Advanced Blocking](./docs/phase3.md) - **✅ COMPLETED**

- [x] The TYPE command ................................................ 🟩
- [x] Create a stream (XADD) .......................................... 🟨
- [x] Validating entry IDs ............................................ 🟥
- [x] Partially auto-generate IDs ..................................... 🟨
- [x] Fully auto-generate IDs ......................................... 🟨
- [x] Query entries into stream (XRANGE) .............................. 🟨
- [x] Query with - .................................................... 🟨
- [x] Query with + .................................................... 🟨
- [x] Query single stream using XREAD ................................. 🟥
- [x] Query multiple streams using XREAD .............................. 🟥
- [x] Blocking reads with timeout ..................................... 🟥
- [x] Blocking reads without timeout (BLOCK 0) ........................ 🟥
- [x] Blocking reads using $ .......................................... 🟥

### [Phase 4: Transactions](./docs/phase4.md) - **COMPLETED**

- [x] The INCR command ................................................ 🟩
- [x] The INCRBY command .............................................. 🟨
- [x] The DECR command ................................................ 🟨
- [x] The DECRBY command .............................................. 🟨
- [x] The MULTI command ............................................... 🟨
- [x] The EXEC command ................................................ 🟥
- [x] Empty transaction ............................................... 🟨
- [x] Queueing commands ............................................... 🟥
- [x] Executing a transaction ......................................... 🟥
- [x] The DISCARD command ............................................. 🟨
- [x] Failures within transactions .................................... 🟥
- [x] Multiple transactions ........................................... 🟥
- [x] Undo Single/Multiple transactions ............................... 🟨

### [Phase 5: Replication](./docs/phase5.md) - **IN PROGRESS**

- [x] Configure listening port ........................................ 🟩
- [x] The INFO command on a replica ................................... 🟩
- [x] The INFO command ................................................ 🟨
- [x] Initial replication ID and offset ............................... 🟩
- [x] Send handshake(1/3) ............................................. 🟩
- [x] Send handshake(2/3) ............................................. 🟩
- [x] Send handshake(3/3) ............................................. 🟨
- [x] Recieve handshake(1/2) .......................................... 🟩
- [x] Receive handshake(2/2)............................................ 🟩
- [x] Empty RDB transfer............................................... 🟩
- [ ] Single-replica propagation ...................................... 🟨
- [ ] Multi-replica propagation ....................................... 🟥
- [ ] Command Processing .............................................. 🟥
- [ ] ACKs with no commands ........................................... 🟩
- [ ] ACKs with commands .............................................. 🟨
- [ ] WAIT with no replicas ........................................... 🟨
- [ ] WAIT with no commands ........................................... 🟨
- [ ] WAIT with multiple commands ..................................... 🟥

### [Phase 6: RDB Persistance](./docs/phase6.md) - **REMAINING**

- [ ] RDB file Config ................................................. 🟩
- [ ] Read a key ...................................................... 🟨
- [ ] Read a string value ............................................. 🟨
- [ ] Read a multiple keys ............................................ 🟨
- [ ] Read multiple string values ..................................... 🟨
- [ ] Read value with expiry .......................................... 🟨

### [Phase 7: PUB/SUB ](./docs/phase7.md) - **REMAINING**

- [ ] Subscribe to multiple channels .................................. 🟩
- [ ] Subscribe to a channel .......................................... 🟩
- [ ] Enter subscribed mode ........................................... 🟨
- [ ] PING in subscribed mode ......................................... 🟩
- [ ] Publish a message ............................................... 🟩
- [ ] Deliver message ................................................. 🟥
- [ ] Unsubscribe ..................................................... 🟨

### [Phase 8: Sorted Sets ](./docs/phase8.md) - **REMAINING**

- [ ] Create a sorted set .................................. 🟩
- [ ] Add members ................................. 🟨
- [ ] Retrieve member rank .................................. 🟨
- [ ] List sorted set members .................................. 🟩
- [ ] ZRANGE with negative indexes .................................. 🟩
- [ ] Count sorted set members .................................. 🟩
- [ ] Retrieve member score .................................. 🟨
- [ ] Remove a member .................................. 🟩

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
│   ├── basic.go                   # PING, ECHO commands
│   ├── dispatch.go                # Command routing & distribution
│   ├── list_blocking.go           # BLPOP, BRPOP commands
│   ├── lists.go                   # LPUSH, RPUSH, LRANGE, LLEN, etc.
│   ├── stream_blocking.go         # XREAD (blocking) commands
│   ├── streams.go                 # XADD, XRANGE commands
│   ├── strings.go                 # SET, GET commands
    ├── transactions.go            # INCR,DECR, MULTI, EXEC, UNDO  commands
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

### **Phase 3: Streams & Advanced Blocking**

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

### **Phase 4: Transactions & Discard**

```bash
# Basic INCR command functionality
127.0.0.1:6379> INCR counter
(integer) 1
127.0.0.1:6379> INCR counter
(integer) 2
127.0.0.1:6379> GET counter
"2"

# INCR with existing numeric value
127.0.0.1:6379> SET mynum "42"
OK
127.0.0.1:6379> INCR mynum
(integer) 43
127.0.0.1:6379> INCR mynum
(integer) 44

# INCR error handling - non-numeric values
127.0.0.1:6379> SET mystring "hello"
OK
127.0.0.1:6379> INCR mystring
(error) ERR value is not an integer or out of range

# INCR error handling - wrong number of arguments
127.0.0.1:6379> INCR
(error) ERR wrong number of arguments for 'incr' command
127.0.0.1:6379> INCR key1 key2
(error) ERR wrong number of arguments for 'incr' command

# INCR overflow protection
127.0.0.1:6379> SET maxint "9223372036854775807"
OK
127.0.0.1:6379> INCR maxint
(error) ERR increment or decrement would overflow

# Empty transaction
127.0.0.1:6379> MULTI
OK
127.0.0.1:6379(TX)> EXEC
(empty array)

# DISCARD command - cancel transaction
127.0.0.1:6379> SET existing "before"
OK
127.0.0.1:6379> MULTI
OK
127.0.0.1:6379(TX)> SET existing "during"
QUEUED
127.0.0.1:6379(TX)> SET new "value"
QUEUED
127.0.0.1:6379(TX)> DISCARD
OK
127.0.0.1:6379> GET existing
"before"
127.0.0.1:6379> GET new
(nil)

# Nested MULTI error handling
127.0.0.1:6379> MULTI
OK
127.0.0.1:6379(TX)> MULTI
(error) ERR MULTI calls can not be nested

# EXEC without MULTI error
127.0.0.1:6379> EXEC
(error) ERR EXEC without MULTI

# DISCARD without MULTI error
127.0.0.1:6379> DISCARD
(error) ERR DISCARD without MULTI

# Transaction with mixed data types
127.0.0.1:6379> MULTI
OK
127.0.0.1:6379(TX)> SET stringkey "hello"
QUEUED
127.0.0.1:6379(TX)> LPUSH listkey "item1" "item2"
QUEUED
127.0.0.1:6379(TX)> INCR numkey
QUEUED
127.0.0.1:6379(TX)> XADD streamkey * field "value"
QUEUED
127.0.0.1:6379(TX)> LLEN listkey
QUEUED
127.0.0.1:6379(TX)> EXEC
1) OK
2) (integer) 2
3) (integer) 1
4) "1754967420156-0"
5) (integer) 2


# Transaction with list operations
127.0.0.1:6379> MULTI
OK
127.0.0.1:6379(TX)> LPUSH mylist "first"
QUEUED
127.0.0.1:6379(TX)> RPUSH mylist "last"
QUEUED
127.0.0.1:6379(TX)> LRANGE mylist 0 -1
QUEUED
127.0.0.1:6379(TX)> LPOP mylist
QUEUED
127.0.0.1:6379(TX)> LLEN mylist
QUEUED
127.0.0.1:6379(TX)> EXEC
1) (integer) 1
2) (integer) 2
3) 1) "first"
   2) "last"
4) "first"
5) (integer) 1

# Transaction with stream operations
127.0.0.1:6379> MULTI
OK
127.0.0.1:6379(TX)> XADD txstream * event "start"
QUEUED
127.0.0.1:6379(TX)> XADD txstream * event "progress"
QUEUED
127.0.0.1:6379(TX)> XRANGE txstream - +
QUEUED
127.0.0.1:6379(TX)> EXEC
1) "1754967430123-0"
2) "1754967430124-0"
3) 1) 1) "1754967430123-0"
      2) 1) "event"
         2) "start"
   2) 1) "1754967430124-0"
      2) 1) "event"
         2) "progress"

# Multiple concurrent transactions (test with multiple terminals)
# Terminal 1:
127.0.0.1:6379> MULTI
OK
127.0.0.1:6379(TX)> SET client1 "data1"
QUEUED
127.0.0.1:6379(TX)> INCR shared_counter
QUEUED

# Terminal 2 (simultaneously):
127.0.0.1:6379> MULTI
OK
127.0.0.1:6379(TX)> SET client2 "data2"
QUEUED
127.0.0.1:6379(TX)> INCR shared_counter
QUEUED

# Terminal 1:
127.0.0.1:6379(TX)> EXEC
1) OK
2) (integer) 1

# Terminal 2:
127.0.0.1:6379(TX)> EXEC
1) OK
2) (integer) 2

# Verify both transactions executed independently:
127.0.0.1:6379> GET client1
"data1"
127.0.0.1:6379> GET client2
"data2"
127.0.0.1:6379> GET shared_counter
"2"

# Transaction with TTL operations
127.0.0.1:6379> MULTI
OK
127.0.0.1:6379(TX)> SET temp1 "value1" EX 60
QUEUED
127.0.0.1:6379(TX)> SET temp2 "value2"
QUEUED
127.0.0.1:6379(TX)> GET temp1
QUEUED
127.0.0.1:6379(TX)> EXEC
1) OK
2) OK
3) "value1"

# Complex transaction with error recovery
127.0.0.1:6379> MULTI
OK
127.0.0.1:6379(TX)> SET key1 "100"
QUEUED
127.0.0.1:6379(TX)> INCR key1
QUEUED
127.0.0.1:6379(TX)> SET key2 "abc"
QUEUED
127.0.0.1:6379(TX)> INCR key2
QUEUED
127.0.0.1:6379(TX)> SET key3 "200"
QUEUED
127.0.0.1:6379(TX)> INCR key3
QUEUED
127.0.0.1:6379(TX)> EXEC
1) OK
2) (integer) 101
3) OK
4) (error) ERR value is not an integer or out of range
5) OK
6) (integer) 201

# Transaction atomicity verification
127.0.0.1:6379> SET balance "1000"
OK
127.0.0.1:6379> MULTI
OK
127.0.0.1:6379(TX)> INCR balance
QUEUED
127.0.0.1:6379(TX)> INCR balance
QUEUED
127.0.0.1:6379(TX)> INCR balance
QUEUED
127.0.0.1:6379(TX)> EXEC
1) (integer) 1001
2) (integer) 1002
3) (integer) 1003
127.0.0.1:6379> GET balance
"1003"

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
