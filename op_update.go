package mongoproto

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mongodb/mongo-tools/common/bsonutil"
	"gopkg.in/mgo.v2/bson"
)

const (
	OpUpdateUpsert OpUpdateFlags = 1 << iota
	OpUpdateMuli
)

type OpUpdateFlags int32

// OpUpdate is used to update a document in a collection.
// http://docs.mongodb.org/meta-driver/latest/legacy/mongodb-wire-protocol/#op-update
type OpUpdate struct {
	Header             MsgHeader
	FullCollectionName string // "dbname.collectionname"
	Flags              OpUpdateFlags
	Selector           []byte // the query to select the document
	Update             []byte // specification of the update to perform
}

func (op *OpUpdate) OpCode() OpCode {
	return OpCodeUpdate
}

func (op *OpUpdate) String() string {
	var doc interface{}
	if err := bson.Unmarshal(op.Selector, &doc); err != nil {
		return "(error unmarshalling selector data)"
	}
	selectorJsonDoc, err := bsonutil.ConvertBSONValueToJSON(doc)
	if err != nil {
		return fmt.Sprintf("ConvertBSONValueToJSON err: %#v - %v", op, err)
	}
	selectorAsJSON, err := json.Marshal(selectorJsonDoc)
	if err != nil {
		return fmt.Sprintf("json marshal err: %#v - %v", op, err)
	}
	if err := bson.Unmarshal(op.Update, &doc); err != nil {
		return "(error unmarshalling update data)"
	}
	updateJsonDoc, err := bsonutil.ConvertBSONValueToJSON(doc)
	if err != nil {
		return fmt.Sprintf("ConvertBSONValueToJSON err: %#v - %v", op, err)
	}
	updateAsJSON, err := json.Marshal(updateJsonDoc)
	if err != nil {
		return fmt.Sprintf("json marshal err: %#v - %v", op, err)
	}
	return fmt.Sprintf("OpUpdate %v %v %v", op.FullCollectionName, string(selectorAsJSON), string(updateAsJSON))
}

func (op *OpUpdate) FromReader(r io.Reader) error {
	var b [4]byte
	_, err := io.ReadFull(r, b[:])
	if err != nil {
		return err
	}
	op.Flags = OpUpdateFlags(getInt32(b[:], 0))
	name, err := readCStringFromReader(r)
	if err != nil {
		return err
	}
	op.FullCollectionName = string(name)
	op.Selector, err = ReadDocument(r)
	if err != nil {
		return err
	}
	var selectorExtra []byte
	selectorExtra, err = ReadDocument(r)
	if err != nil {
		return err
	}
	op.Selector = append(op.Selector, selectorExtra...)
	op.Update, err = ReadDocument(r)
	if err != nil {
		return err
	}
	return nil
}
