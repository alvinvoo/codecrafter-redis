package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/util"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30)

		reader := bufio.NewReader(tcpConn)
		// Set a read deadline of 2 seconds for each command
		tcpConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		for {
			request, err := reader.ReadString('\n') // will block (stuck) until the delimiter is found
			util.DebugLog("Request", request)

			if err == io.EOF {
				util.DebugLog("Connection closed by client")
				break
			}
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				util.DebugLog("Timeout on reading request")
				break
			}
			// catch all error
			if err != nil {
				util.DebugLog("Error reading request", err)
				break
			}

			// ignore RESP prefix first
			if request[0] == '*' || request[0] == '$' {
				continue
			}

			response := "+PONG\r\n"

			_, err = tcpConn.Write([]byte(response))
			if err != nil {
				fmt.Println("Error writing response: ", err.Error())
				os.Exit(1)
			}
		}
	}
}
