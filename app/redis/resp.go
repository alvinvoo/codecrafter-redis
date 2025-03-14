package redis

import (
	"slices"
	"strconv"
	"strings"
)

func stripNewlines(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '\r' || s[i] == '\n' {
			s = strings.Replace(s, "\r", " ", -1)
			s = strings.Replace(s, "\n", " ", -1)
			break
		}
	}
	return s
}

func convertIntToBytes(n int) []byte {
	var retB []byte
	for n != 0 {
		l := byte('0' + n%10)

		retB = append(retB, l)
		n = n / 10
	}

	slices.Reverse(retB)
	return retB
}

// appendPrefix will append a "$3\r\n" style redis prefix for a message.
func appendPrefix(b []byte, c byte, n int) []byte {
	b = append(b, c)
	b = strconv.AppendInt(b, int64(n), 10)
	return append(b, '\r', '\n')
}

// AppendError appends a Redis protocol error to the input bytes.
func AppendError(b []byte, s string) []byte {
	b = append(b, '-')
	b = append(b, stripNewlines(s)...)
	return append(b, '\r', '\n')
}

// AppendOK appends a Redis protocol OK to the input bytes.
func AppendOK(b []byte) []byte {
	return append(b, '+', 'O', 'K', '\r', '\n')
}

func AppendArray(b []byte, msgs []string) []byte {
	b = append(b, '*')
	b = append(b, convertIntToBytes(len(msgs))...)
	b = append(b, '\r', '\n')

	for _, msg := range msgs {
		b = AppendBulkString(b, msg)
	}

	return b
}

// AppendString appends a Redis protocol string to the input bytes.
func AppendString(b []byte, s string) []byte {
	b = append(b, '+')
	b = append(b, stripNewlines(s)...)
	return append(b, '\r', '\n')
}

func AppendBulk(b []byte, bulk []byte) []byte {
	b = appendPrefix(b, '$', len(bulk))
	b = append(b, bulk...)
	return append(b, '\r', '\n')
}

// AppendBulkString appends a Redis protocol bulk string to the input bytes.
func AppendBulkString(b []byte, bulk string) []byte {
	b = appendPrefix(b, '$', len(bulk))
	b = append(b, bulk...)
	return append(b, '\r', '\n')
}

func AppendNull(b []byte) []byte {
	return append(b, '$', '-', '1', '\r', '\n')
}
