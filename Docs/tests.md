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

# Advanced Blocking Features

# Blocking XREAD with timeout (test with multiple terminals)
# Terminal 1:
127.0.0.1:6379> XREAD BLOCK 5000 STREAMS livestream $
# (waits for up to 5 seconds for new data)

# Terminal 2 (while Terminal 1 is waiting):
127.0.0.1:6379> XADD livestream * event "real-time" data "immediate"
"1754967414819-0"

# Terminal 1 immediately receives:
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
```

### **Phase 4: Transactions**

```bash
# Basic INCR command functionality
127.0.0.1:6379> INCR counter
(integer) 1
127.0.0.1:6379> INCR counter
(integer) 2

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
127.0.0.1:6379(TX)> DISCARD
OK
127.0.0.1:6379> GET existing
"before"

# Transaction with mixed operations
127.0.0.1:6379> MULTI
OK
127.0.0.1:6379(TX)> SET key1 "100"
QUEUED
127.0.0.1:6379(TX)> INCR key1
QUEUED
127.0.0.1:6379(TX)> LPUSH mylist "item"
QUEUED
127.0.0.1:6379(TX)> EXEC
1) OK
2) (integer) 101
3) (integer) 1

# UNDO command - remove queued commands
127.0.0.1:6379> MULTI
OK
127.0.0.1:6379(TX)> SET key1 "val1"
QUEUED
127.0.0.1:6379(TX)> SET key2 "val2"
QUEUED
127.0.0.1:6379(TX)> SET key3 "val3"
QUEUED
127.0.0.1:6379(TX)> UNDO 2
*4
$22
Removed 2 commands:
$13
SET key2 val2
$13
SET key3 val3
$29
1 commands remaining in queue
```

### **Phase 5: Replication**

```bash
# Check replication status on master
127.0.0.1:6379> INFO replication
# Replication
role:master
connected_slaves:1
master_replid:a1b2c3d4e5f6...
master_repl_offset:0

# Write operations are automatically propagated to replicas
127.0.0.1:6379> SET mykey "value"
OK
127.0.0.1:6379> INCR counter
(integer) 1

# WAIT for replica synchronization
127.0.0.1:6379> WAIT 1 1000
(integer) 1  # Number of replicas that acknowledged

# Check replication on slave (port 6380)
127.0.0.1:6380> INFO replication
# Replication
role:slave
master_host:localhost
master_port:6379
master_link_status:up

# Verify data was replicated
127.0.0.1:6380> GET mykey
"value"
127.0.0.1:6380> GET counter
"1"
```
