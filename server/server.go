package server

import (
	"fmt"
	"net"

	"github.com/kushalsdesk/redis_with_go/server/handler"
)

func ListenAndServe(addr string) {

	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Failed to bind to ", addr)
		return
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error Accepting Connection:", err)
			continue
		}

		fmt.Println("Accepting new Connection:", conn.RemoteAddr())
		go handler.HandleConnection(conn)
	}
}
