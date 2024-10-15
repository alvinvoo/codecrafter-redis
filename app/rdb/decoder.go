package rdb

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/util"
)

const (
	rdb6bitLen  = 0
	rdb14bitLen = 1
	rdb32bitLen = 2
	rdbEncVal   = 3

	rdbFlagMeta     = 0xfa
	rdbFlagResizeDB = 0xfb
	rdbFlagExpiryMS = 0xfc
	rdbFlagExpiryS  = 0xfd
	rdbFlagSelectDB = 0xfe
	rdbFlagEOF      = 0xff

	rdbEncInt8  = 0
	rdbEncInt16 = 1
	rdbEncInt32 = 2
)

type Decoder interface {
	StartRDB()

	StartDatabase(n uint32)

	Meta(key, value []byte)

	Set(key, value []byte, expiry int64)

	EndDatabase(n uint32)

	EndRDB()
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
	d.event.StartRDB()

	var objType byte
	var key, value []byte
	var expiry int64
	var db uint32

	for {

		objType, err = d.r.ReadByte()
		if err != nil {
			return err
		}

		switch objType {
		case rdbFlagMeta:
			key, err = d.readString()
			if err != nil {
				return err
			}

			value, err = d.readString()
			if err != nil {
				return err
			}

			d.event.Meta(key, value)

		case rdbFlagResizeDB:
			htSize, _, err := d.readLength()
			if err != nil {
				return err
			}

			htExpirySize, _, err := d.readLength()
			if err != nil {
				return err
			}

			log.Printf("resizeDB hash table size %d expiry size %d\n", htSize, htExpirySize)
		case rdbFlagExpiryMS:
			// intBuf is 8 bytes as initiliazed in Decoder
			// but just to make it clear
			_, err := io.ReadFull(d.r, d.intBuf[:8])
			if err != nil {
				return err
			}
			expiry = int64(binary.LittleEndian.Uint64(d.intBuf))
		case rdbFlagExpiryS:
			_, err := io.ReadFull(d.r, d.intBuf[:4])
			if err != nil {
				return err
			}
			expiry = int64(binary.LittleEndian.Uint32(d.intBuf)) * 1000
		case rdbFlagSelectDB:
			db, _, err = d.readLength()
			if err != nil {
				return err
			}

			d.event.StartDatabase(db)
		case rdbFlagEOF:
			d.event.EndDatabase(db)
			d.event.EndRDB()
			return nil
		default: // default here is "00"
			key, err = d.readString()
			if err != nil {
				return err
			}

			value, err = d.readString()
			if err != nil {
				return err
			}

			d.event.Set(key, value, expiry)
		}
	}
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

// max is 32 bit
func (d *decode) readLength() (uint32, bool, error) {
	b, err := d.r.ReadByte()
	if err != nil {
		return 0, false, err
	}

	switch (b & 0xC0) >> 6 { // the bitwise AND part is redundant, but i guess its to improve readability
	case rdb6bitLen:
		return uint32(b & 0x3f), false, nil
	case rdb14bitLen:
		bb, err := d.r.ReadByte()
		if err != nil {
			return 0, false, err
		}

		// big endian
		return (uint32(b&0x3f)<<8 | uint32(bb)), false, nil
	case rdb32bitLen:
		_, err := io.ReadFull(d.r, d.intBuf[:4])
		if err != nil {
			return 0, false, err
		}

		return binary.BigEndian.Uint32(d.intBuf[:4]), false, nil
	case rdbEncVal:
		return uint32(b & 0x3f), true, nil
	default:
		return 0, false, fmt.Errorf("length encoding not supported")
	}
}

func (d *decode) readString() ([]byte, error) {
	length, encode, err := d.readLength()
	if err != nil {
		return nil, err
	}

	if encode {
		switch length {
		case rdbEncInt8:
			i, err := d.readUint8()
			return []byte(strconv.Itoa(int(i))), err
		case rdbEncInt16:
			i, err := d.readUint16()
			return []byte(strconv.Itoa(int(i))), err
		case rdbEncInt32:
			i, err := d.readUint32()
			return []byte(strconv.Itoa(int(i))), err
		default:
			return nil, fmt.Errorf("string encoding not supported")
		}
	}

	str := make([]byte, length)
	_, err = io.ReadFull(d.r, str)
	return str, err
}

func (d *decode) readUint8() (uint8, error) {
	b, err := d.r.ReadByte()
	return uint8(b), err
}

func (d *decode) readUint16() (uint16, error) {
	_, err := io.ReadFull(d.r, d.intBuf[:2])
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(d.intBuf[:2]), nil
}

func (d *decode) readUint32() (uint32, error) {
	_, err := io.ReadFull(d.r, d.intBuf[:4])
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(d.intBuf[:4]), nil
}

// Decode parses a RDB file from r and calls the decode hooks on d.
func Decode(r io.Reader, d Decoder) error {
	decoder := &decode{d, make([]byte, 8), bufio.NewReader(r)}
	return decoder.decode()
}
