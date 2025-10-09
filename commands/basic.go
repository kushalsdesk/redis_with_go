package commands

import (
	"fmt"
	"net"
	"strings"
	"time"

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

func handleConfig(args []string, conn net.Conn) {
	if len(args) < 2 {
		conn.Write([]byte("-ERR wrong number of arguments for 'config' command\r\n"))
		return
	}

	subcommand := strings.ToUpper(args[1])

	switch subcommand {
	case "GET":
		handleConfigGet(args, conn)
	case "SET":
		conn.Write([]byte("-ERR CONFIG SET is not supported\r\n"))
	default:
		conn.Write([]byte(fmt.Sprintf("-ERR unknown CONFIG subcommand '%s'\r\n", subcommand)))
	}
}

func handleConfigGet(args []string, conn net.Conn) {

	if len(args) != 3 {
		conn.Write([]byte("-ERR wrong number of arguments for 'config get' command \r\n"))
		return
	}

	parameter := strings.ToLower(args[2])

	value, exists := store.GetConfigValue(parameter)

	if !exists {
		conn.Write([]byte("*0\r\n"))
		return
	}

	resp := fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
		len(parameter), parameter,
		len(value), value)
	conn.Write([]byte(resp))

}

func handleInfo(args []string, conn net.Conn) {
	replState := store.GetReplicationState()
	section := ""

	if len(args) > 1 {
		section = strings.ToLower(args[1])
	}

	var info strings.Builder

	if section == "" || section == "replication" {
		info.WriteString("# Replication\r\n")

		if replState.Role == "master" {
			info.WriteString("role:master\r\n")
			info.WriteString(fmt.Sprintf("connected_slaves:%d\r\n", replState.ConnectedSlaves))
			info.WriteString(fmt.Sprintf("master_replid:%s\r\n", replState.MasterReplID))
			info.WriteString(fmt.Sprintf("master_repl_offset:%d\r\n", replState.MasterReplOffset))

			minOffset := int64(-1)
			conns := store.GetReplicaConnections()
			for _, rep := range conns {
				if minOffset == -1 || rep.Offset < minOffset {
					minOffset = rep.Offset
				}
			}
			info.WriteString(fmt.Sprintf("second_repl_offset:%d\r\n", minOffset))

			info.WriteString("repl_backlog_active:0\r\n")
			info.WriteString("repl_backlog_size:1048576\r\n")
			info.WriteString("repl_backlog_first_byte_offset:0\r\n")
			info.WriteString("repl_backlog_histlen:0\r\n")

			for i, rep := range conns {
				linkStatus := "online"
				if time.Since(rep.LastACK) > 10*time.Second {
					linkStatus = "disconnected"
				}
				info.WriteString(fmt.Sprintf("slave%d:ip=%s,port=...,state=%s,offset=%d,lag=%d\r\n",
					i, rep.Address, linkStatus, rep.Offset, rep.Lag))
			}
		} else {
			info.WriteString("role:slave\r\n")
			info.WriteString(fmt.Sprintf("master_host:%s\r\n", replState.MasterHost))
			info.WriteString(fmt.Sprintf("master_port:%s\r\n", replState.MasterPort))

			linkStatus := "down"
			if store.GetSlaveOffset() > 0 {
				linkStatus = "up"
			}
			lastIO := -1
			if linkStatus == "up" {
				lastIO = 1
			}
			info.WriteString(fmt.Sprintf("master_link_status:%s\r\n", linkStatus))
			info.WriteString(fmt.Sprintf("master_last_io_seconds_ago:%d\r\n", lastIO))
			info.WriteString("master_sync_in_progress:0\r\n")
			info.WriteString(fmt.Sprintf("slave_repl_offset:%d\r\n", store.GetSlaveOffset())) // Dynamic
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
