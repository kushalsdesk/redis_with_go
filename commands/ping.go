package commands

import "net"

func handlePing(conn net.Conn) {
	conn.Write([]byte("+PONG\r\n"))

}
