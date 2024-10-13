package rdb

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/util"
)

type Decoder interface {
	StartDatabase(n int)

	Meta(key, value []byte)

	Set(key, value []byte, expiry int64)
}

type byteReader interface {
	io.Reader // satisfy io.ReadFull interface
	io.ByteReader
}

type decode struct {
	event  Decoder
	intBuf []byte
	r      byteReader
}

func (d *decode) decode() error {
	err := d.checkHeader()
	if err != nil {
		return err
	}

	return nil
}

func (d *decode) checkHeader() error {
	header := make([]byte, 9)
	_, err := io.ReadFull(d.r, header)
	if err != nil {
		return err
	}

	if !bytes.Equal(header[:5], []byte("REDIS")) {
		return fmt.Errorf("rdb: invalid file magic format")
	}

	version, _ := strconv.ParseInt(string(header[5:]), 10, 64)
	if version < 1 || version > 11 {
		return fmt.Errorf("rdb: invalid RDB version number %d", version)
	}

	util.DebugLog("RDB version: ", version)

	return nil
}

// Decode parses a RDB file from r and calls the decode hooks on d.
func Decode(r io.Reader, d Decoder) error {
	decoder := &decode{d, make([]byte, 8), bufio.NewReader(r)}
	return decoder.decode()
}
