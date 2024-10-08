package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/redis"
)

const addr = ":6379"

func main() {
	go log.Printf("started server at %s", addr)

	items := make(map[string][]byte)
	mu := sync.RWMutex{}

	err := redis.ListenAndServeTLS(addr,
		func(conn redis.Conn, cmd redis.Command) {
			cmdTxt := strings.ToLower(string(cmd.Args[0]))
			switch cmdTxt {
			default:
				conn.WriteError(fmt.Sprintf("Error unknwon command: %s", cmdTxt))
			case "ping":
				conn.WriteString("PONG")
			case "echo":
				if len(cmd.Args) != 2 {
					conn.WriteError("Invalid ECHO arg length")
					return
				}

				conn.WriteBulk(cmd.Args[1])
			case "set":
				if len(cmd.Args) != 3 {
					conn.WriteError("SET command needs 2 arg")
					return
				}

				mu.Lock()
				items[string(cmd.Args[1])] = cmd.Args[2]
				mu.Unlock()

				conn.WriteOK()
			case "get":
				if len(cmd.Args) != 2 {
					conn.WriteError("Invalid GET arg length")
					return
				}

				mu.RLock()
				val, ok := items[string(cmd.Args[1])]
				mu.RUnlock()

				if !ok {
					conn.WriteNull()
				} else {
					conn.WriteBulk(val)
				}
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
