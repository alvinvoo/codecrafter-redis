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

// func handleConnection(conn net.Conn) {
// 	defer conn.Close()
// 	if tcpConn, ok := conn.(*net.TCPConn); ok {
// 		tcpConn.SetKeepAlive(true)
// 		tcpConn.SetKeepAlivePeriod(30)

// 		reader := bufio.NewReader(tcpConn)
// 		// Set a read deadline of 2 seconds for each command
// 		tcpConn.SetReadDeadline(time.Now().Add(2 * time.Second))
// 		for {
// 			request, err := reader.ReadString('\n') // will block (stuck) until the delimiter is found
// 			util.DebugLog("Request", request)

// 			if err == io.EOF {
// 				util.DebugLog("Connection closed by client")
// 				break
// 			}
// 			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
// 				util.DebugLog("Timeout on reading request")
// 				break
// 			}
// 			// catch all error
// 			if err != nil {
// 				util.DebugLog("Error reading request", err)
// 				break
// 			}

// 			// ignore RESP prefix first
// 			if request[0] == '*' || request[0] == '$' {
// 				continue
// 			}

// 			response := "+PONG\r\n"

// 			_, err = tcpConn.Write([]byte(response))
// 			if err != nil {
// 				fmt.Println("Error writing response: ", err.Error())
// 				os.Exit(1)
// 			}
// 		}
// 	}
// }
