package mongoproto

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrInvalidSize = errors.New("mongoproto: got invalid document size")
)

const (
	maximumDocumentSize = 16 * 1024 * 1024 // 16MB max
)

// ReadDocument read an entire BSON document. This document can be used with
// bson.Unmarshal.
func ReadDocument(r io.Reader) ([]byte, error) {
	var sizeRaw [4]byte
	if _, err := io.ReadFull(r, sizeRaw[:]); err != nil {
		return nil, err
	}
	size := getInt32(sizeRaw[:], 0)
	if size < 0 {
		return nil, ErrInvalidSize
	}
	if size > maximumDocumentSize {
		return nil, ErrInvalidSize
	}
	doc := make([]byte, size)
	if size == 0 {
		return doc, nil
	}
	if size < 4 {
		return doc, nil
	}
	setInt32(doc, 0, size)

	if _, err := io.ReadFull(r, doc[4:]); err != nil {
		return doc, err
	}
	return doc, maybeCheckBSON(doc)
}

// readCStringFromReader reads a null turminated string from an io.Reader.
func readCStringFromReader(r io.Reader) ([]byte, error) {
	var b []byte
	var n [1]byte
	for {
		if _, err := io.ReadFull(r, n[:]); err != nil {
			return nil, err
		}
		if n[0] == 0 {
			return b, nil
		}
		b = append(b, n[0])
	}
}

// all data in the MongoDB wire protocol is little-endian.
// all the read/write functions below are little-endian.

func getInt32(b []byte, pos int) int32 {
	return (int32(b[pos+0])) |
		(int32(b[pos+1]) << 8) |
		(int32(b[pos+2]) << 16) |
		(int32(b[pos+3]) << 24)
}

func setInt32(b []byte, pos int, i int32) {
	b[pos] = byte(i)
	b[pos+1] = byte(i >> 8)
	b[pos+2] = byte(i >> 16)
	b[pos+3] = byte(i >> 24)
}

func getInt64(b []byte, pos int) int64 {
	return (int64(b[pos+0])) |
		(int64(b[pos+1]) << 8) |
		(int64(b[pos+2]) << 16) |
		(int64(b[pos+3]) << 24) |
		(int64(b[pos+4]) << 32) |
		(int64(b[pos+5]) << 40) |
		(int64(b[pos+6]) << 48) |
		(int64(b[pos+7]) << 56)
}

type leWriter struct {
	w   io.Writer
	err error
}

func (l *leWriter) Write(data interface{}) error {
	if l.err != nil {
		return l.err
	}
	l.err = binary.Write(l.w, binary.LittleEndian, data)
	return l.err
}
