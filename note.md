## Notes
- Clients send commands to the Redis server as RESP arrays.
- that's y PING start with asterisk *
- e.g. : *1\r\n$4\r\nPING\r\n
- Array format: `*<number-of-elements>\r\n<element-1>...<element-n>`
- Bulk string format: `$<length>\r\n<data>\r\n`



## Qns
### 7th Oct 2024
1. echo -e "PING" | redis-cli
PONG
This works but its actually sending 2 commands in non-iterative mode
1. COMMAND DOCS
2. PING
The command docs first hit and then _i/o time out_ before another socket is opened to send the ping command

BUT, the `Redcon` code is able to process all commands within one socket connection, and there are actually 3 commands in total
1. COMMAND DOCS
2. COMMAND
3. PING

Need to check what's the difference and how is it able to maintain the connection without using `c.conn.SetReadDeadline` 


guess1: Is it becoz i never move the start end? That's y it get's "Stucked" if there's no timeout?
*2\r\n$7\r\nCOMMAND\r\n$4\r\nDOCS\r\n*1\r\n$7\r\nCOMMAND\r\n

Investigation bits:
bufio io.Reader.Read
// Read reads data into p.
// It returns the number of bytes read into p.
// The bytes are taken from at most one Read on the underlying [Reader],
// hence n may be less than len(p).

Every ONE Read is actually just one command section
No idea what the "underlying [Reader]", cant find it on code, (is it directly on assembly level?)
 - Ans: Its based on https://pkg.go.dev/net#TCPConn.Read
 TCPConn
** Implicit qns is, HOW does it know when to _stop_? (as percisely one command section?)

ONCE the net.TCPConn _writes_ the next set of bytes is _ready_ in the recv-Q
- i think due to the Protocol is waiting for RESP reply?
- SHOULD BE due to client is waiting for response before sending again?
 https://redis.io/docs/latest/develop/reference/protocol-spec/#request-response-model

 ### Useful
 To check opened sockets
 `ss -s`
  - get socket summaries
 `ss -t -a`
  - filter by TCP, VERY useful to see RECV and SEND Queue
  - "Everything in Linux is a file"
    - https://en.wikipedia.org/wiki/Everything_is_a_file
    - that means data stream are handled as "file descriptors"