package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/redis"
)

const addr = ":6379"

func main() {
	go log.Printf("started server at %s", addr)

	err := redis.ListenAndServeTLS(addr,
		func(conn redis.Conn, cmd redis.Command) {
			cmdTxt := strings.ToLower(string(cmd.Args[0]))
			switch cmdTxt {
			default:
				conn.WriteError(fmt.Sprintf("Error unknwon command: %s", cmdTxt))
			case "ping":
				conn.WriteString("PONG")
			case "echo":
				reply := cmd.Args[1]
				if len(reply) == 0 {
					conn.WriteError("Invalid ECHO arg length")
				}

				conn.WriteBulk(reply)
			}
		},
		func(conn redis.Conn) bool {
			log.Printf("accept: %s", conn.RemoteAddr())
			return true
		},
		func(conn redis.Conn, err error) {
			log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
		})

	if err != nil {
		log.Fatal(err)
	}
}
