After building image 1.1 after step 1
󱓞  podman ps
CONTAINER ID IMAGE COMMAND CREATED STATUS PORTS NAMES
94c79def49f7 6 seconds ago Up 6 seconds 0.0.0.0:16379-16380->6379-6380/tcp 0c7a6e17a3e1-infra
e39b514fe9c7 localhost/redis-clone:v1.1 ./redis-clone 6 seconds ago Up 6 seconds 0.0.0.0:16379-16380->6379-6380/tcp redis-pod-master
4c9d6891bb0f localhost/redis-clone:v1.1 6 seconds ago Up 5 seconds 0.0.0.0:16379-16380->6379-6380/tcp redis-pod-slave

redis_with_go  devel  ×1  ×2 via  v1.24.7
󱓞  podman logs redis-pod-master
Starting Redis server as master on 0.0.0.0:6379Accepting new Connection: 10.89.0.7:36020
🔗 Replica connected: 10.89.0.7:36020
📦 Sent empty RDB (88 bytes) to replica

redis_with_go  devel  ×1  ×2 via  v1.24.7
󱓞  podman logs redis-pod-slave
Starting Redis server as replica of master 6379 on 0.0.0.0:6380 🚀 Starting replication with master master:6379
🔗 Connected to master master:6379
✅ PING successful
✅ REPLCONF listening-port successful
✅ REPLCONF capa successful
✅ PSYNC successful: +FULLRESYNC f53c3b4ad82df31746abb982d8e9dd2a1748abc5 0
📦 Received RDB file (88 bytes)
📋 RDB version: 0011
✅ RDB validation successful
🎉 Replication handshake completed!
📡 Listening for propagated commands...

redis_with_go  devel  ×1  ×2 via  v1.24.7
󱓞  podman logs redis-pod-master
Starting Redis server as master on 0.0.0.0:6379Accepting new Connection: 10.89.0.7:36020
🔗 Replica connected: 10.89.0.7:36020
📦 Sent empty RDB (88 bytes) to replica
Accepting new Connection: 10.89.0.7:35280
📡 Propagating to 1 replicas: [set mekey value1] (size ~94 bytes)
📊 Master offset updated: +94 (total: 94)
✅ Command propagated successfully
📡 Propagating to 1 replicas: [lpush llist ele mene bene] (size ~141 bytes)
📊 Master offset updated: +141 (total: 235)
✅ Command propagated successfully
📡 Propagating to 1 replicas: [incr counter] (size ~71 bytes)
📊 Master offset updated: +71 (total: 306)
✅ Command propagated successfully
📡 Propagating to 1 replicas: [incrby counter 2] (size ~94 bytes)
📊 Master offset updated: +94 (total: 400)
✅ Command propagated successfully

redis_with_go  devel  ×1  ×2 via  v1.24.7
󱓞  podman logs redis-pod-slave
Starting Redis server as replica of master 6379 on 0.0.0.0:6380 🚀 Starting replication with master master:6379
🔗 Connected to master master:6379
✅ PING successful
✅ REPLCONF listening-port successful
✅ REPLCONF capa successful
✅ PSYNC successful: +FULLRESYNC f53c3b4ad82df31746abb982d8e9dd2a1748abc5 0
📦 Received RDB file (88 bytes)
📋 RDB version: 0011
✅ RDB validation successful
🎉 Replication handshake completed!
📡 Listening for propagated commands...
Accepting new Connection: 10.89.0.7:47330
📥 Received: [set mekey value1]
✅ Replicated SET mekey = value1
📊 Slave offset updated: 94
📥 Received: [lpush llist ele mene bene]
✅ Replicated LPUSH llist (length: 3)
📊 Slave offset updated: 235
📥 Received: [incrby counter 2]
✅ Replicated INCRBY counter 2 = 2
📊 Slave offset updated: 329
📥 Received: [incr counter]
✅ Replicated INCR counter = 3
📊 Slave offset updated: 400

After Changing with Global ACK channel with Enchanced handleReplConf

redis_with_go  devel  ×5  x1  ×2 via  v1.24.7
󱓞  podman logs redis-pod-master
Starting Redis server as master on 0.0.0.0:6379Accepting new Connection: 10.89.0.10:42526
🔗 Replica connected: 10.89.0.10:42526
📦 Sent empty RDB (88 bytes) to replica
📊Updated replica 10.89.0.10:42526: offset=0, lag=0
📊Updated replica 10.89.0.10:42526: offset=0, lag=0
📊Updated replica 10.89.0.10:42526: offset=0, lag=0
📊Updated replica 10.89.0.10:42526: offset=0, lag=0
📊Updated replica 10.89.0.10:42526: offset=0, lag=0
📊Updated replica 10.89.0.10:42526: offset=0, lag=0
📊Updated replica 10.89.0.10:42526: offset=0, lag=0
📊Updated replica 10.89.0.10:42526: offset=0, lag=0
📊Updated replica 10.89.0.10:42526: offset=0, lag=0
📊Updated replica 10.89.0.10:42526: offset=0, lag=0

redis_with_go  devel  ×5  x1  ×2 via  v1.24.7
󱓞  podman logs redis-pod-slave
📢 Initialized ACK channel for slave
Starting Redis server as replica of master 6379 on 0.0.0.0:6380 🚀 Starting replication with master master:6379
🔗 Connected to master master:6379
✅ PING successful
✅ REPLCONF listening-port successful
✅ REPLCONF capa successful
✅ PSYNC successful: +FULLRESYNC 96577b6ce8f2582d779652fd45b16b7285fc1443 0
📦 Received RDB file (88 bytes)
📋 RDB version: 0011
✅ RDB validation successful
🎉 Replication handshake completed!
📡 Listening for propagated commands...
📤 Slave sent ACK to master: offset=0
📤 Slave sent ACK to master: offset=0
📤 Slave sent ACK to master: offset=0
📤 Slave sent ACK to master: offset=0
📤 Slave sent ACK to master: offset=0
📤 Slave sent ACK to master: offset=0
📤 Slave sent ACK to master: offset=0
📤 Slave sent ACK to master: offset=0
📤 Slave sent ACK to master: offset=0
📤 Slave sent ACK to master: offset=0
📤 Slave sent ACK to master: offset=0
📤 Slave sent ACK to master: offset=0
📤 Slave sent ACK to master: offset=0
📤 Slave sent ACK to master: offset=0
