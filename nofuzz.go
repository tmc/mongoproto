// +build !gofuzz

package mongoproto

import (
	"bytes"
	"io"
	"log"
)

func maybeCheckBSON(b []byte) error {
	return nil
}

// yet-to be used methods to get more accurate fuzz coverage data

func (op *OpGetMore) fromWire(b []byte) {
	b = b[4:] // skip ZERO
	op.FullCollectionName = readCString(b)
	b = b[len(op.FullCollectionName)+1:]
	op.NumberToReturn = getInt32(b, 0)
	op.CursorID = getInt64(b, 4)
}

func (op *OpGetMore) toWire() []byte {
	return nil
}

func (op *OpReply) fromWire(b []byte) {
	if len(b) < 20 {
		return
	}
	op.Flags = OpReplyFlags(getInt32(b, 0))
	op.CursorID = getInt64(b, 4)
	op.StartingFrom = getInt32(b, 12)
	op.NumberReturned = getInt32(b, 16)

	offset := 20
	for i := int32(0); i < op.NumberReturned; i++ {
		doc, err := ReadDocument(bytes.NewReader(b[offset:]))
		if err != nil {
			// TODO(tmc) probably should return an error from fromWire
			log.Println("doc err:", err, len(b[offset:]))
			break
		}
		op.Documents = append(op.Documents, doc)
		offset += len(doc)
	}
}

func (op *OpQuery) toWire() []byte {
	return nil
}

func (op *OpReply) toWire() []byte {
	return nil
}

func (op *OpInsert) toWire() []byte {
	return nil
}

func (op *OpUnknown) fromWire(b []byte) {
}

func (op *OpUnknown) toWire() []byte {
	return nil
}

func readCString(b []byte) string {
	for i := 0; i < len(b); i++ {
		if b[i] == 0 {
			return string(b[:i])
		}
	}
	return ""
}

// IsMutation tells us if the operation will mutate data. These operations can
// be followed up by a getLastErr operation.
func (c OpCode) IsMutation() bool {
	return c == OpCodeInsert || c == OpCodeUpdate || c == OpCodeDelete
}

// HasResponse tells us if the operation will have a response from the server.
func (c OpCode) HasResponse() bool {
	return c == OpCodeQuery || c == OpCodeGetMore
}

// CopyMessage copies reads & writes an entire message.
func CopyMessage(w io.Writer, r io.Reader) error {
	h, err := ReadHeader(r)
	if err != nil {
		return err
	}
	if _, err := h.WriteTo(w); err != nil {
		return err
	}
	_, err = io.CopyN(w, r, int64(h.MessageLength-MsgHeaderLen))
	return err
}
