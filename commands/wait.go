package commands

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handleWait(args []string, conn net.Conn) {
	if len(args) != 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'wait' command\r\n"))
		return
	}

	numReplicaStr, timeoutStr := args[1], args[2]
	numReplicas, err := strconv.Atoi(numReplicaStr)

	if err != nil || numReplicas < 0 {
		conn.Write([]byte("-ERR invalid first argument. The number of replica must be >= 0\r\n"))
		return
	}
	timeoutMs, err := strconv.Atoi(timeoutStr)
	if err != nil || timeoutMs < 0 {
		conn.Write([]byte("-ERR invalid second argument. The timeout must be >= 0\r\n"))
		return
	}

	totalRep := store.GetNumReplicas()
	if numReplicas > totalRep {
		resp := fmt.Sprintf(":%d\r\n", 0)
		conn.Write([]byte(resp))
		return
	}

	if totalRep == 0 {
		resp := fmt.Sprintf(":%d\r\n", 0)
		conn.Write([]byte(resp))
		return
	}

	targetOffset := store.GetReplOffset()

	if targetOffset == 0 {
		resp := fmt.Sprintf(":%d\r\n", totalRep)
		conn.Write([]byte(resp))
		return
	}
	acked := waitForACKs(numReplicas, timeoutMs, targetOffset)
	resp := fmt.Sprintf(":%d\r\n", acked)
	conn.Write([]byte(resp))
}

func waitForACKs(numReplicas, timeoutMs int, targetOffset int64) int {
	start := time.Now()
	ackedCount := 0
	for ackedCount < numReplicas {
		if time.Since(start) > time.Duration(timeoutMs)*time.Millisecond {
			fmt.Printf("⏰ WAIT timeout after %d ms, %d/%d replicas ACKed (target offset %d)\n", timeoutMs, ackedCount, numReplicas, targetOffset)
			return ackedCount
		}

		lags := store.GetAllReplicaLags()
		currentAcked := 0
		for _, lag := range lags {
			if lag <= 0 {
				currentAcked++
			}
		}
		ackedCount = currentAcked

		if ackedCount >= numReplicas {
			fmt.Printf("✅ WAIT succeeded: %d replicas ACKed (target %d)\n", ackedCount, targetOffset)
			return ackedCount
		}

		time.Sleep(50 * time.Millisecond)
	}
	return ackedCount
}
