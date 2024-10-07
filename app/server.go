package main

import (
	"log"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/redis"
)

const addr = ":6379"

func main() {
	go log.Printf("started server at %s", addr)

	err := redis.ListenAndServeTLS(addr,
		func(conn net.Conn, cmd redis.Command) {
			cmdTxt := strings.ToLower(string(cmd.Args[0]))
			switch cmdTxt {
			default:
				response := "-error\r\n"

				_, err := conn.Write([]byte(response))
				if err != nil {
					log.Println("Error writing response: ", err.Error())
					os.Exit(1)
				}

				log.Printf("Error unknown command: %s \n", cmdTxt)
			case "ping":
				response := "+PONG\r\n"

				_, err := conn.Write([]byte(response))
				if err != nil {
					log.Println("Error writing response: ", err.Error())
					os.Exit(1)
				}
			}
		},
		func(conn net.Conn) bool {
			log.Printf("accept: %s", conn.RemoteAddr())
			return true
		},
		func(conn net.Conn, err error) {
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
