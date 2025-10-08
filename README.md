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

### [Phase 4: Transactions](./docs/phase4.md) - **✅ COMPLETED**

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

### [Phase 5: Replication](./docs/phase5.md) - **✅ COMPLETED**

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
- [x] Single-replica propagation ...................................... 🟨
- [x] Multi-replica propagation ....................................... 🟥
- [x] Command Processing .............................................. 🟥
- [x] ACKs with no commands ........................................... 🟩
- [x] ACKs with commands .............................................. 🟨
- [x] WAIT with no replicas ........................................... 🟨
- [x] WAIT with no commands ........................................... 🟨
- [x] WAIT with multiple commands ..................................... 🟥

### [Phase 6: RDB Persistance](./docs/phase6.md) - **⏳ REMAINING**

- [ ] RDB file Config ................................................. 🟩
- [ ] Read a key ...................................................... 🟨
- [ ] Read a string value ............................................. 🟨
- [ ] Read a multiple keys ............................................ 🟨
- [ ] Read multiple string values ..................................... 🟨
- [ ] Read value with expiry .......................................... 🟨

### [Phase 7: PUB/SUB ](./docs/phase7.md) - **⏳ REMAINING**

- [ ] Subscribe to multiple channels .................................. 🟩
- [ ] Subscribe to a channel .......................................... 🟩
- [ ] Enter subscribed mode ........................................... 🟨
- [ ] PING in subscribed mode ......................................... 🟩
- [ ] Publish a message ............................................... 🟩
- [ ] Deliver message ................................................. 🟥
- [ ] Unsubscribe ..................................................... 🟨

### [Phase 8: Sorted Sets ](./docs/phase8.md) - **⏳ REMAINING**

- [ ] Create a sorted set ............................................. 🟩
- [ ] Add members ..................................................... 🟨
- [ ] Retrieve member rank ............................................ 🟨
- [ ] List sorted set members ......................................... 🟩
- [ ] ZRANGE with negative indexes .................................... 🟩
- [ ] Count sorted set members ........................................ 🟩
- [ ] Retrieve member score ........................................... 🟨
- [ ] Remove a member ................................................. 🟩

## Project Structure

```
redis_with_go/
├── app/
│   └── main.go                       # Application entry point with CLI flags & server initialization
│
├── server/
│   ├── server.go                     # TCP server setup and connection acceptance
│   ├── replication.go                # Replication client logic (handshake, RDB transfer, command sync)
│   └── handler/
│       └── handler.go                # RESP protocol parsing & connection lifecycle management
│
├── commands/                         # Command handlers and business logic
│   ├── dispatch.go                   # Command routing, transaction detection & replication propagation
│   ├── basic.go                      # PING, ECHO, INFO commands
│   ├── strings.go                    # SET, GET commands with TTL support
│   ├── counter.go                    # INCR, DECR, INCRBY, DECRBY atomic operations
│   ├── lists.go                      # LPUSH, RPUSH, LPOP, RPOP, LRANGE, LLEN, LINDEX
│   ├── list_blocking.go              # BLPOP, BRPOP with timeout/infinite blocking support
│   ├── streams.go                    # XADD (with ID validation/generation), XRANGE
│   ├── stream_blocking.go            # XREAD with BLOCK support and $ handling
│   ├── transactions.go               # MULTI, EXEC, DISCARD, UNDO transaction management
│   ├── replication.go                # PSYNC, REPLCONF (listening-port, capa, ACK) handlers
│   ├── propagation.go                # Write command detection & RESP encoding for replication
│   ├── wait.go                       # WAIT command for replica synchronization
│   └── utils.go                      # TYPE command for key type inspection
│
├── store/                            # Data storage layer with concurrency control
│   ├── core.go                       # Core data structures (RedisValue, Stream, ReplicationState)
│   │                                 # Key type detection, expiry checking, replica management
│   ├── string_ops.go                 # String storage (Set, Get, Delete) with TTL
│   │                                 # Counter operations (Increment, Decrement with overflow protection)
│   ├── list_ops.go                   # List operations (Push, Pop, Range, Index, Length)
│   ├── list_blocking.go              # Blocking client registration, notification system for lists
│   ├── stream_ops.go                 # Stream storage (Add, Range, ReadFrom)
│   │                                 # ID parsing, validation, generation (auto/partial)
│   ├── stream_blocking.go            # Blocking client registration, notification system for streams
│   └── replication.go                # Replication offset tracking, ACK management
│                                     # Replica lag calculation, command size estimation
│
├── docs/
│   ├── phase1.md                     # Phase 1 implementation details
│   ├── phase2.md                     # Phase 2 implementation details
│   ├── phase3.md                     # Phase 3 implementation details
│   ├── phase4.md                     # Phase 4 implementation details
│   └── phase5.md                     # Phase 5 implementation details
│
├── Dockerfile                        # Multi-stage Docker build for production deployment
├── redis-cluster.yaml                # Kubernetes pod spec for master-slave setup
├── go.mod                            # Go module dependencies
└── README.md                         # This file
```

## Architecture Highlights

### 🏗️ **Layered Architecture**
- **Server Layer**: TCP connection handling and RESP protocol parsing
- **Command Layer**: Business logic and command execution
- **Store Layer**: Thread-safe data storage and replication state

### 🔒 **Concurrency & Safety**
- Mutex-protected data structures for concurrent client access
- Separate read/write locks for optimal performance
- Channel-based notification system for blocking operations

### 🔄 **Replication System**
- Full master-slave replication with PSYNC protocol
- Automatic command propagation to replicas
- ACK-based synchronization with lag tracking
- WAIT command for ensuring replica consistency

### 🎯 **Blocking Operations**
- Event-driven architecture using Go channels
- Client registration system for BLPOP/BRPOP/XREAD
- Timeout support (finite and infinite blocking)
- Immediate notification when data becomes available

### 💾 **Transaction Support**
- Command queueing with MULTI/EXEC
- Per-connection transaction state
- DISCARD for cancellation
- Custom UNDO command for removing queued commands

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

### 2. Run as Master-Slave Cluster

```bash
# Terminal 1 - Start master
go run app/main.go --port 6379

# Terminal 2 - Start slave
go run app/main.go --port 6380 --replicaof "localhost 6379"
```

### 3. Connect with Redis CLI

```bash
redis-cli -p 6379
```

### 4. Docker Deployment

```bash
# Build image
docker build -t redis-clone:latest .

# Run container
docker run -p 6379:6379 redis-clone:latest
```


## Performance Features

### 🚀 **Real-time Notifications**
- Immediate response when data arrives for blocking commands
- Event-driven architecture using Go channels
- No polling overhead

### 💪 **Concurrent Client Support**
- Multiple redis-cli connections work simultaneously
- Independent blocking on different keys/streams
- Automatic client cleanup on disconnection

### ⚡ **Memory Efficiency**
- Modular storage with separated concerns
- TTL support with automatic cleanup
- Type-safe data structures

### 🔄 **Replication Efficiency**
- Command propagation with size estimation
- ACK-based tracking for lag monitoring
- Efficient RESP encoding for network transfer

## Key Learning Outcomes

- **TCP Server Programming**: Building concurrent network servers in Go
- **Protocol Implementation**: Redis RESP (Redis Serialization Protocol)
- **Data Structures**: Implementing Redis data types (strings, lists, streams)
- **Concurrency**: Goroutines, channels, mutexes, and thread-safe operations
- **Memory Management**: TTL implementation and expiration handling
- **Advanced Patterns**: Real-time blocking operations, event-driven architecture
- **Distributed Systems**: Master-slave replication, command propagation, consistency
- **Software Architecture**: Clean code principles, modular design, separation of concerns

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available under the MIT License.
