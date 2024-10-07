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


- second time is lesser, 4069, coz minus 27 bytes


problem FD
*internal/poll.FD {fdmu: internal/poll.fdMutex {state: 10, rsema: 0, wsema: 0}, Sysfd: 5, SysFile: internal/poll.SysFile {iovecs: *[]syscall.Iovec nil}, pd: internal/poll.pollDesc {runtimeCtx: 139728111398280}, csema: 0, isBlocking: 0, IsStream: true, ZeroReadIsEOF: true, isFile: false}


EWOULDBLOCK (11) = 0xb
 - Something else (another thread) has drained the input buffer
 - A receive timeout was set on the socket and it expired without data being received
 - your thread would have to block in order to do that 


*internal/poll.FD {fdmu: internal/poll.fdMutex {state: 10, rsema: 0, wsema: 0}, Sysfd: 5, SysFile: internal/poll.SysFile {iovecs: *[]syscall.Iovec nil}, pd: internal/poll.pollDesc {runtimeCtx: 140282157743448}, csema: 0, isBlocking: 0, IsStream: true, ZeroReadIsEOF: true, isFile: false}


LISTEN          0               4096                                        *:6380                                        *:*              ESTAB           27              0                          [::ffff:127.0.0.1]:6380                       [::ffff:127.0.0.1]:46402


ONCE the net.TCPConn _writes_ the next set of bytes is _ready_ in the recv-Q
-- i think due to the Protocol is waiting for RESP reply?
 https://redis.io/docs/latest/develop/reference/protocol-spec/#request-response-model