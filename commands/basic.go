package commands

import (
	"fmt"
	"net"
	"strings"

	"github.com/kushalsdesk/redis_with_go/store"
)

func handlePing(conn net.Conn) {
	conn.Write([]byte("+PONG\r\n"))
}

func handleEcho(args []string, conn net.Conn) {
	if len(args) < 2 {
		conn.Write([]byte("-ERR wrong number of arguments\r\n"))
		return
	}
	msg := args[1]
	resp := fmt.Sprintf("$%d\r\n%s\r\n", len(msg), msg)
	conn.Write([]byte(resp))
}

func handleInfo(args []string, conn net.Conn) {
	// Get current replication state
	replState := store.GetReplicationState()

	var section string
	if len(args) > 1 {
		section = strings.ToLower(args[1])
	}

	var info strings.Builder

	if section == "" || section == "replication" {
		info.WriteString("# Replication\r\n")

		if replState.Role == "master" {
			// Master role information
			info.WriteString("role:master\r\n")
			info.WriteString(fmt.Sprintf("connected_slaves:%d\r\n", replState.ConnectedSlaves))
			info.WriteString(fmt.Sprintf("master_replid:%s\r\n", replState.MasterReplID))
			info.WriteString(fmt.Sprintf("master_repl_offset:%d\r\n", replState.MasterReplOffset))
			info.WriteString("second_repl_offset:-1\r\n")
			info.WriteString("repl_backlog_active:0\r\n")
			info.WriteString("repl_backlog_size:1048576\r\n")
			info.WriteString("repl_backlog_first_byte_offset:0\r\n")
			info.WriteString("repl_backlog_histlen:0\r\n")
		} else {
			// Replica role information
			info.WriteString("role:slave\r\n")
			info.WriteString(fmt.Sprintf("master_host:%s\r\n", replState.MasterHost))
			info.WriteString(fmt.Sprintf("master_port:%s\r\n", replState.MasterPort))
			info.WriteString("master_link_status:down\r\n") // Will be "up" after handshake
			info.WriteString("master_last_io_seconds_ago:-1\r\n")
			info.WriteString("master_sync_in_progress:0\r\n")
			info.WriteString("slave_repl_offset:0\r\n")
			info.WriteString("slave_priority:100\r\n")
			info.WriteString("slave_read_only:1\r\n")
		}
	}

	if section == "" || section == "server" {
		if info.Len() > 0 {
			info.WriteString("\r\n")
		}
		info.WriteString("# Server\r\n")
		info.WriteString("redis_version:7.0.0\r\n")
		info.WriteString("redis_git_sha1:00000000\r\n")
		info.WriteString("redis_git_dirty:0\r\n")
		info.WriteString("redis_build_id:0\r\n")

		if replState.Role == "master" {
			info.WriteString("redis_mode:standalone\r\n")
		} else {
			info.WriteString("redis_mode:slave\r\n")
		}

		info.WriteString("os:Linux\r\n")
		info.WriteString("arch_bits:64\r\n")
		info.WriteString("multiplexing_api:epoll\r\n")
		info.WriteString("gcc_version:0.0.0\r\n")
		info.WriteString("process_id:1\r\n")
		info.WriteString("tcp_port:6379\r\n")
		info.WriteString("uptime_in_seconds:1\r\n")
		info.WriteString("uptime_in_days:0\r\n")
	}

	infoStr := info.String()
	resp := fmt.Sprintf("$%d\r\n%s\r\n", len(infoStr), infoStr)
	conn.Write([]byte(resp))
}
