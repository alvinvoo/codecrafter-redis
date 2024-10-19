package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/rdb"
	"github.com/codecrafters-io/redis-starter-go/app/redis"
)

const addr = ":6379"

func maybeFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	go log.Printf("started server at %s", addr)

	mu := sync.RWMutex{}

	if len(os.Args) > 1 {
		redis.SetConfig(os.Args[1:])
		dbFilePath, err := redis.GetRdbPath()
		maybeFatal(err)

		f, err := os.Open(dbFilePath)
		maybeFatal(err)

		err = rdb.Decode(f, &rdb.StoreWorker{})
		maybeFatal(err)
	}

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
				if len(cmd.Args) < 3 {
					conn.WriteError("SET command needs at least 2 arg")
					return
				}

				var expiry int64 = 0
				if len(cmd.Args) > 3 {
					// process options
					switch strings.ToLower(string(cmd.Args[3])) {
					default:
						conn.WriteError("SET arg not supported")
					case "px":
						if len(cmd.Args) != 5 {
							conn.WriteError("SET arg px needs duration")
						}

						var err error
						expiry, err = strconv.ParseInt(string(cmd.Args[4]), 10, 64)

						if err != nil {
							conn.WriteError(fmt.Sprintf("SET arg px parse expiry error %s", err.Error()))
						}
					}
				}

				mu.Lock()
				err := rdb.StoreRDB(string(cmd.Args[1]), cmd.Args[2], expiry)
				maybeFatal(err)
				mu.Unlock()

				conn.WriteOK()
			case "get":
				if len(cmd.Args) != 2 {
					conn.WriteError("Invalid GET arg length")
					return
				}

				mu.RLock()
				data, ok := rdb.GetRDB(string(cmd.Args[1]))
				mu.RUnlock()

				if !ok {
					conn.WriteNull()
				} else {
					conn.WriteBulk(data)
				}
			case "keys":
				if len(cmd.Args) != 2 {
					conn.WriteError("Invalid KEYS arg length")
					return
				}

				if string(cmd.Args[1]) != "*" {
					conn.WriteError("Only support KEYS * for now")
					return
				}

				mu.RLock()
				keys := rdb.GetRDBKeys()
				if len(keys) == 0 {
					conn.WriteNull()
				} else {
					conn.WriteArray(keys)
				}
				mu.RUnlock()
			case "config":
				if len(cmd.Args) != 3 {
					conn.WriteError("Invalid CONFIG arg length")
					return
				}

				action := strings.ToLower(string(cmd.Args[1]))

				if action != "get" {
					conn.WriteError("Only CONFIG GET supported for now")
					return
				}

				switch strings.ToLower(string(cmd.Args[2])) {
				default:
					conn.WriteError("Unrecognized config key")
				case "dir":
					config := redis.GetConfig()
					conn.WriteArray([]string{
						"dir", config.Dir,
					})
				case "dbfilename":
					config := redis.GetConfig()
					conn.WriteArray([]string{
						"dbfilename", config.Dbfilename,
					})
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

	maybeFatal(err)
}
