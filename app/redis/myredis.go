package redis

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const BUFSIZE = 4096 // default buffer size in Readers

var (
	errInvalidBulkLength      = &errProtocol{"invalid bulk length"}
	errInvalidMultiBulkLength = &errProtocol{"invalid multibulk length"}
)

type errProtocol struct {
	msg string
}

func (e *errProtocol) Error() string {
	return "Protocol error: " + e.msg
}

// Writer allows for writing RESP messages.
type Writer struct {
	w io.Writer
	b []byte
}

// NewWriter creates a new RESP writer.
func NewWriter(wr io.Writer) *Writer {
	return &Writer{
		w: wr,
	}
}

func (wr *Writer) Flush() error {
	// TODO: do we need to store `err` in Writer?
	_, err := wr.w.Write(wr.b)
	wr.b = nil
	return err
}

func (wr *Writer) WriteError(msg string) {
	wr.b = AppendError(wr.b, msg)
}

func (wr *Writer) WriteArray(msgs []string) {
	wr.b = AppendArray(wr.b, msgs)
}

func (wr *Writer) WriteString(msg string) {
	wr.b = AppendString(wr.b, msg)
}

func (wr *Writer) WriteBulk(msg []byte) {
	wr.b = AppendBulk(wr.b, msg)
}

func (wr *Writer) WriteBulkString(msg string) {
	wr.b = AppendBulkString(wr.b, msg)
}

func (wr *Writer) WriteNull() {
	wr.b = AppendNull(wr.b)
}

func (wr *Writer) WriteOK() {
	wr.b = AppendOK(wr.b)
}

type Command struct {
	Raw  []byte
	Args [][]byte // each command can have multiple args
	// like "echo 2", "hget 10"
}

type Reader struct {
	rd    *bufio.Reader
	buf   []byte
	start int
	end   int
	// cmds  []Command // there could be multiple commands
}

func NewReader(rd io.Reader) *Reader {
	return &Reader{
		rd:  bufio.NewReader(rd),
		buf: make([]byte, BUFSIZE),
	}
}

func (rd *Reader) readCommands() ([]Command, error) {
	var cmds []Command
	// happy path, with no leftovers
	// no idea why; but each read would be one full command
	// we need to continue reading to capture all the commands sent in one connection
	n, err := rd.rd.Read(rd.buf[rd.end:])
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}

	rd.end += n
	// start is needed here
	b := rd.buf[rd.start:rd.end]
	switch b[0] {
	case '*':
		// resp formatted command

		marks := make([]int, 0, 16)

		bufLen := n
		for i := 1; i < bufLen; i++ {
			if b[i] == '\n' {
				if b[i-1] != '\r' {
					return nil, errInvalidMultiBulkLength
				}
				// read array count
				count, err := strconv.Atoi(string(b[1 : i-1]))
				if err != nil {
					return nil, err
				}
				if count <= 0 {
					return nil, errInvalidMultiBulkLength
				}
				for j := 0; j < count; j++ {
					// read bulk string length
					i++
					if i < bufLen {
						if b[i] != '$' {
							return nil, &errProtocol{"expected '$', got '" +
								string(b[i]) + "'"}
						}
						si := i
						for ; i < bufLen; i++ {
							if b[i] == '\n' {
								if b[i-1] != '\r' {
									return nil, errInvalidBulkLength
								}
								size, err := strconv.Atoi(string(b[si+1 : i-1]))
								if err != nil {
									return nil, errInvalidBulkLength
								}
								if b[i+size+2] != '\n' || b[i+size+1] != '\r' {
									return nil, errInvalidBulkLength
								}
								i++
								marks = append(marks, i, i+size)
								i += size + 1
								break
							}
						}
					}
				}
				if len(marks) == count*2 {
					var cmd Command

					cmd.Raw = append([]byte(nil), b[:i+1]...)

					cmd.Args = make([][]byte, len(marks)/2)

					for h := 0; h < len(marks); h += 2 {
						cmd.Args[h/2] = cmd.Raw[marks[h]:marks[h+1]]
					}

					cmds = append(cmds, cmd)
				}
			}
		}

		rd.start = rd.end
		return cmds, nil
	default:
		return nil, fmt.Errorf("unexpected character: %c", rd.buf[0])
	}

}
