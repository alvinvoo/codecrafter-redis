package redis

import (
	"fmt"
	"io"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/util"
)

type conn struct {
	conn net.Conn
	rd   *Reader
	wr   *Writer
}

type Server struct {
	net     string
	laddr   string
	handler func(conn net.Conn, cmd Command)
	accept  func(conn net.Conn) bool
	closed  func(conn net.Conn, err error)
	conn    *conn // KISS; only ONE connection for now
	ln      net.Listener
}

// ListenAndServeTLS creates a new TLS server and binds to addr configured on "tcp" network net.
func ListenAndServeTLS(addr string,
	handler func(conn net.Conn, cmd Command),
	accept func(conn net.Conn) bool,
	closed func(conn net.Conn, err error),
) error {
	return NewServerNetwork("tcp", addr, handler, accept, closed).ListenAndServe()
}

func NewServerNetwork(
	net, laddr string,
	handler func(conn net.Conn, cmd Command),
	accept func(conn net.Conn) bool,
	closed func(conn net.Conn, err error),
) *Server {
	if handler == nil {
		panic("handler is nil")
	}
	s := &Server{}
	s.net = net
	s.laddr = laddr
	s.handler = handler
	s.accept = accept
	s.closed = closed
	return s
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen(s.net, s.laddr)
	if err != nil {
		return fmt.Errorf("failed to bind to %s", s.laddr)
	}
	s.ln = ln

	return s.serve()
}

func (s *Server) serve() error {
	defer func() {
		s.ln.Close()
	}()

	for {
		lnconn, err := s.ln.Accept()
		if err != nil {
			return fmt.Errorf("error accepting connection: %s", err.Error())
		}

		c := &conn{
			conn: lnconn,
			rd:   NewReader(lnconn),
			wr:   NewWriter(lnconn),
		}

		s.conn = c

		s.accept(c.conn)

		go s.handle()
	}
}

func (s *Server) handle() {
	var err error
	c := s.conn
	defer func() {
		// check errors if needed
		c.conn.Close()

		if s.closed != nil {
			if err == io.EOF { // if EOF, just ignore
				err = nil
			}
			s.closed(c.conn, err)
		}
	}()

	err = func() error {
		for {
			// should not need to set read timeout
			// c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			// readCommands is called multiple times, until EOF

			// with our current request-response scenario
			// there should only be ONE command at a time
			cmds, err := c.rd.readCommands()
			if err != nil {
				return err
			}

			for _, cm := range cmds {
				util.DebugLog("Got commands: ", cm)
				for _, a := range cm.Args {
					util.DebugLog("with args", string(a))
				}
			}

			for len(cmds) > 0 {
				cmd := cmds[0]
				if len(cmds) == 1 {
					cmds = nil
				} else {
					cmds = cmds[1:]
				}
				s.handler(c.conn, cmd)
			}
		}
	}()
}
