package mongoproto

import (
	"fmt"
	"io"

	"github.com/mongodb/mongo-tools/common/bsonutil"
	"github.com/mongodb/mongo-tools/common/json"
	"gopkg.in/mgo.v2/bson"
)

const (
	OpDeleteSingleRemove OpDeleteFlags = 1 << iota
)

type OpDeleteFlags int32

// OpDelete is used to remove one or more documents from a collection.
// http://docs.mongodb.org/meta-driver/latest/legacy/mongodb-wire-protocol/#op-delete
type OpDelete struct {
	Header             MsgHeader
	FullCollectionName string // "dbname.collectionname"
	Flags              OpDeleteFlags
	Selector           []byte // the query to select the document(s)
}

func (op *OpDelete) String() string {
	var query interface{}
	if err := bson.Unmarshal(op.Selector, &query); err != nil {
		return "(error unmarshalling)"
	}
	queryAsJSON, err := bsonutil.ConvertBSONValueToJSON(query)
	if err != nil {
		return fmt.Sprintf("ConvertBSONValueToJSON err: %#v - %v", op, err)
	}
	asJSON, err := json.Marshal(queryAsJSON)
	if err != nil {
		return fmt.Sprintf("json marshal err: %#v - %v", op, err)
	}
	return fmt.Sprintf("OpDelete %v %v", op.FullCollectionName, string(asJSON))
}

func (op *OpDelete) OpCode() OpCode {
	return OpCodeDelete
}

func (op *OpDelete) FromReader(r io.Reader) error {
	var b [4]byte
	_, err := io.ReadFull(r, b[:])
	if err != nil {
		return err
	}
	op.Flags = OpDeleteFlags(getInt32(b[:], 0))
	name, err := readCStringFromReader(r)
	if err != nil {
		return err
	}
	op.FullCollectionName = string(name)
	op.Selector, err = ReadDocument(r)
	if err != nil {
		return err
	}
	if int(op.Header.MessageLength) > len(op.Selector) + len(op.FullCollectionName) + 1 + 8 + MsgHeaderLen {
		data, err := ReadDocument(r)
		if err != nil {
			return err
		}
		op.Selector = append(op.Selector, data...)
	}
	return nil
}
