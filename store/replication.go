package store

import (
	"fmt"
	"net"
	"time"
)

// Global channel for ACK triggers
var ackOffsetChan chan int64

func UpdateMasterOffset(delta int64) {
	IncrementReplOffset(delta)
	fmt.Printf("📊 Master offset updated: +%d (total: %d)\n", delta, GetReplOffset())
}

func UpdateReplicaOffset(replicaID string, offset int64) {
	replicationMutex.Lock()
	defer replicationMutex.Unlock()

	if rep, exists := replicationState.ReplicaConns[replicaID]; exists && rep.Connected {
		rep.Offset = offset
		rep.LastACK = time.Now()
		rep.Lag = replicationState.MasterReplOffset - offset
		fmt.Printf("📊Updated replica %s: offset=%d, lag=%d\n", replicaID, offset, rep.Lag)
	}
}

func GetSlaveOffset() int64 {
	replicationMutex.RLock()
	defer replicationMutex.RUnlock()
	return replicationState.SlaveOffset
}

func SetSlaveOffset(newOffset int64) {
	replicationMutex.Lock()
	defer replicationMutex.Unlock()
	replicationState.SlaveOffset = newOffset
	fmt.Printf("📊 Slave offset updated: %d\n", newOffset)
}

func GetNumReplicas() int {
	replicationMutex.RLock()
	defer replicationMutex.RUnlock()

	return replicationState.ConnectedSlaves
}

func GetReplicaLag(replicaID string) int64 {
	replicationMutex.RLock()
	defer replicationMutex.RUnlock()

	if rep, exists := replicationState.ReplicaConns[replicaID]; exists && rep.Connected {
		return rep.Lag
	}
	return -1
}

func GetAllReplicaLags() map[string]int64 {
	replicationMutex.RLock()
	defer replicationMutex.RUnlock()

	lags := make(map[string]int64)
	for id, rep := range replicationState.ReplicaConns {
		if rep.Connected {
			lags[id] = rep.Lag
		}
	}
	return lags
}

func EstimateCommandSize(args []string) int64 {
	size := int64(20)
	for _, arg := range args {
		size += int64(len(arg) + 20)
	}
	return size
}

func InitializeReplicaFields(conn net.Conn, replID string) {
	replicationMutex.Lock()
	defer replicationMutex.Unlock()

	id := getReplicaID(conn)
	if rep, exists := replicationState.ReplicaConns[id]; exists {
		rep.ReplID = replID
		rep.Offset = 0
		rep.LastACK = time.Now()
		rep.Lag = replicationState.MasterReplOffset
	}
}

func InitACKChannel() {
	if ackOffsetChan == nil {
		ackOffsetChan = make(chan int64, 10)
		fmt.Printf("📢 Initialized ACK channel for slave\n")
	}
}

func CloseACKChannel() {
	if ackOffsetChan != nil {
		close(ackOffsetChan)
		ackOffsetChan = nil
	}
}

func SendACKTrigger(offset int64) {
	if ackOffsetChan != nil {
		ackOffsetChan <- offset
		fmt.Printf("📢 Triggered immediate ACK for offset %d\n", offset)
	}
}

// GetACKChannel returns the channel for ticker
func GetACKChannel() <-chan int64 {
	return ackOffsetChan
}
