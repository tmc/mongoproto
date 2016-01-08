// +build gofuzz

package mongoproto

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/mgo.v2/bson"
)

type canWrite interface {
	WriteTo(io.Writer) (int64, error)
}

func Fuzz(data []byte) int {
	outBuf := &bytes.Buffer{}
	op, err := OpFromReader(bytes.NewReader(data))
	if err != nil {
		return 0
	}
	op.OpCode()
	fmt.Sprint(op)
	if writer, ok := op.(canWrite); ok {
		if _, err := writer.WriteTo(outBuf); err != nil {
			// TODO(tmc): compare in vs out lengths?
			return 0
		}
	}
	return 1
}

func maybeCheckBSON(b []byte) error {
	var v interface{}
	return bson.Unmarshal(b, &v)
}
