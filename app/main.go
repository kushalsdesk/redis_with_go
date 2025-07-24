package main

import "github.com/kushalsdesk/redis_with_go/server"

func main() {

	server.ListenAndServe("0.0.0.0:6379")
}
