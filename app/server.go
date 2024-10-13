package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/rdb"
	"github.com/codecrafters-io/redis-starter-go/app/redis"
)

const addr = ":6379"

type option struct {
	storedAt  time.Time
	expiresIn time.Duration
}

type store struct {
	data   []byte
	option *option
}

func maybeFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	go log.Printf("started server at %s", addr)

	items := make(map[string]store)
	mu := sync.RWMutex{}

	if len(os.Args) > 1 {
		redis.SetConfig(os.Args[1:])
		dbFilePath, err := redis.GetRdbPath()
		maybeFatal(err)

		f, err := os.Open(dbFilePath)
		maybeFatal(err)

		err = rdb.Decode(f, &rdb.Store{})
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

				var opt *option
				if len(cmd.Args) > 3 {
					// process options
					switch strings.ToLower(string(cmd.Args[3])) {
					default:
						conn.WriteError("SET arg not supported")
					case "px":
						if len(cmd.Args) != 5 {
							conn.WriteError("SET arg px needs duration")
						}

						duration, err := time.ParseDuration(string(cmd.Args[4]) + "ms")

						if err != nil {
							conn.WriteError(fmt.Sprintf("SET arg px parse duration error %s", err.Error()))
						}

						opt = &option{
							storedAt:  time.Now(),
							expiresIn: duration,
						}
					}

				}

				mu.Lock()
				items[string(cmd.Args[1])] = store{
					data:   cmd.Args[2],
					option: opt,
				}
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
					if val.option != nil {
						if (val.option.storedAt.Add(val.option.expiresIn)).Before(time.Now()) {
							delete(items, string(cmd.Args[1]))
							conn.WriteNull()
							return
						}
					}

					conn.WriteBulk(val.data)
				}
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
