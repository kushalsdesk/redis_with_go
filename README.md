# Redis Clone with Go

[![progress-banner](https://backend.codecrafters.io/progress/redis/0a412eea-657f-434d-b2cc-b7352c66c04f)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)

This is a starting point for Go solutions to the ["Build Your Own Redis" Challenge](https://codecrafters.io/challenges/redis).

## Overview

Redis is an in-memory data structure store often used as a database, cache, message broker and streaming engine. In this challenge you'll build your own Redis server that is capable of serving basic commands, reading RDB files and more.

Along the way, we'll learn about TCP servers, the Redis Protocol, data structures, concurrency patterns, and advanced Redis features like transactions and streams.

**Note**: If you're viewing this repo on GitHub, head over to [codecrafters.io](https://codecrafters.io) to try the challenge.

## Implementation Progress

> **Difficulty Levels:** 🟩 Easy | 🟨 Medium | 🟥 Hard

### [Phase 1: Basic Server & String Operations](./docs/phase1.md) - **COMPLETED**

- [x] Bind to a port ................................................... 🟩⬜⬜
- [x] Respond to PING .................................................. 🟩⬜⬜
- [x] Respond to multiple PINGS ........................................ 🟩⬜⬜
- [x] Handle concurrent clients ........................................ 🟩🟨⬜
- [x] Implement the ECHO command ....................................... 🟩⬜⬜
- [x] Implement the SET & GET command .................................. 🟩🟨⬜
- [x] Expiry ........................................................... 🟩🟨⬜

### [Phase 2: Lists](./docs/phase2.md) - **COMPLETED**

- [x] Create a list ................................................... 🟩⬜⬜
- [x] Append an element (RPUSH) ....................................... 🟩⬜⬜
- [x] Append multiple elements ........................................ 🟩⬜⬜
- [x] List elements (positive indexes) ................................ 🟩🟨⬜
- [x] List elements (negative indexes) ................................ 🟩🟨⬜
- [x] Prepend elements (LPUSH) ........................................ 🟩⬜⬜
- [x] Query list length ............................................... 🟩⬜⬜
- [x] Remove an element ............................................... 🟩🟨⬜
- [x] Remove multiple elements ........................................ 🟩🟨⬜
- [x] Blocking retrieval .............................................. 🟩🟨🟥
- [x] Blocking retrieval with timeout ................................. 🟩🟨🟥

### [Phase 3: Streams](./docs/phase3.md) - **COMPLETED**

- [x] The TYPE command ................................................ 🟩⬜⬜
- [x] Create a stream ................................................. 🟩🟨⬜
- [x] Validating entry IDs ............................................ 🟩🟨🟥
- [x] Partially auto-generate IDs ..................................... 🟩🟨⬜
- [x] Fully auto-generate IDs ......................................... 🟩🟨⬜
- [x] Query entries into stream ....................................... 🟩🟨⬜
- [x] Query with - .................................................... 🟩🟨⬜
- [x] Query with + .................................................... 🟩🟨⬜
- [x] Query single stream using XREAD ................................. 🟩🟨🟥
- [x] Query multiple streams using XREAD .............................. 🟩🟨🟥
- [x] Blocking reads .................................................. 🟩🟨🟥
- [x] Blocking reads without timeout .................................. 🟩🟨🟥
- [x] Blocking reads using $ .......................................... 🟩🟨🟥

### [Phase 4: Transactions](./docs/phase4.md) - **NOT STARTED**

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
git clone https://github.com/kushalsdesk/redis_with_go
cd redis_with_go

# Run the server
go run app/main.go

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
