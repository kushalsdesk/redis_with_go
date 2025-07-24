# Redis Clone with Go

[![progress-banner](https://backend.codecrafters.io/progress/redis/0a412eea-657f-434d-b2cc-b7352c66c04f)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)

This is a starting point for Go solutions to the ["Build Your Own Redis" Challenge](https://codecrafters.io/challenges/redis).

## Overview

Redis is an in-memory data structure store often used as a database, cache, message broker and streaming engine. In this challenge you'll build your own Redis server that is capable of serving basic commands, reading RDB files and more.

Along the way, we'll learn about TCP servers, the Redis Protocol, data structures, concurrency patterns, and advanced Redis features like transactions and streams.

**Note**: If you're viewing this repo on GitHub, head over to [codecrafters.io](https://codecrafters.io) to try the challenge.

## Implementation Progress

### [Phase 1: Basic Server & String Operations](./Docs/phase1.md) - **COMPLETED**

- [x] Bind to a port *(Easy)*
- [x] Respond to PING *(Easy)*
- [x] Respond to multiple PINGS *(Easy)*
- [x] Handle concurrent clients *(Medium)*
- [x] Implement the ECHO command *(Easy)*
- [x] Implement the SET & GET command *(Medium)*
- [x] Expiry *(Medium)*

### [Phase 2: Lists](./Docs/phase2.md) - **IN PROGRESS**

- [x] Create a list *(Easy)*
- [x] Append an element (RPUSH) *(Easy)*
- [x] Append multiple elements *(Easy)*
- [ ] List elements (positive indexes) *(Medium)*
- [ ] List elements (negative indexes) *(Medium)*
- [x] Prepend elements (LPUSH) *(Easy)*
- [ ] Query list length *(Easy)*
- [ ] Remove an element *(Medium)*
- [ ] Remove multiple elements *(Medium)*
- [ ] Blocking retrieval *(Hard)*
- [ ] Blocking retrieval with timeout *(Hard)*

### [Phase 3: Streams](./Docs/phase3.md) - **NOT STARTED**

- [ ] The TYPE command *(Easy)*
- [ ] Create a stream *(Medium)*
- [ ] Validating entry IDs *(Hard)*
- [ ] Partially auto-generate IDs *(Medium)*
- [ ] Fully auto-generate IDs *(Medium)*
- [ ] Query entries into stream *(Medium)*
- [ ] Query with - *(Medium)*
- [ ] Query with + *(Medium)*
- [ ] Query single stream using XREAD *(Hard)*
- [ ] Query multiple streams using XREAD *(Hard)*
- [ ] Blocking reads *(Hard)*
- [ ] Blocking reads without timeout *(Hard)*
- [ ] Blocking reads using $ *(Hard)*

### [Phase 4: Transactions](./Docs/phase4.md) - **NOT STARTED**

- [ ] The INCR command (1/3) *(Easy)*
- [ ] The INCR command (2/3) *(Medium)*
- [ ] The INCR command (3/3) *(Medium)*
- [ ] The MULTI command *(Medium)*
- [ ] The EXEC command *(Hard)*
- [ ] Empty transaction *(Medium)*
- [ ] Queueing commands *(Hard)*
- [ ] Executing a transaction *(Hard)*
- [ ] The DISCARD command *(Medium)*
- [ ] Failures within transactions *(Hard)*
- [ ] Multiple transactions *(Hard)*

## Project Structure

```
├── main.go                 # Entry point
├── server/
│   ├── server.go          # TCP server & connection handling
│   └── handler/
│       └── handler.go     # RESP protocol parsing
├── commands/
│   ├── dispatch.go        # Command routing
│   ├── ping.go           # PING command
│   ├── echo.go           # ECHO command
│   ├── set.go            # SET command
│   ├── get.go            # GET command
│   ├── lpush.go          # LPUSH command
│   └── rpush.go          # RPUSH command
├── store/
│   └── memory.go         # In-memory storage with TTL
└── docs/
    ├── phase1.md         # Phase 1 implementation details
    └── phase2.md         # Phase 2 implementation details (coming soon)
```

## Getting Started

```bash
# Clone the repository
git clone <repository-url>
cd redis_with_go

# Run the server
go run main.go

# Test with redis-cli or telnet
redis-cli -p 6379
> PING
PONG
> SET mykey "hello"
OK
> GET mykey
"hello"
> LPUSH mylist "world" "hello"
(integer) 2
```

## Key Learning Outcomes

- **TCP Server Programming**: Building concurrent network servers in Go
- **Protocol Implementation**: Redis RESP (Redis Serialization Protocol)
- **Data Structures**: Implementing Redis data types (strings, lists, streams)
- **Concurrency**: Goroutines, mutexes, and thread-safe operations
- **Memory Management**: TTL implementation and expiration handling
- **Advanced Patterns**: Blocking operations, transactions, and streaming
