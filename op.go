package mongoproto

import (
	"io"
)

// Op is a Mongo operation
type Op interface {
	OpCode() OpCode
	FromReader(io.Reader) error
}

// OpFromReader reads an Op from an io.Reader
func OpFromReader(r io.Reader) (Op, error) {
	msg, err := ReadHeader(r)
	if err != nil {
		return nil, err
	}
	m := *msg

	var result Op
	switch m.OpCode {
	case OpCodeQuery:
		result = &OpQuery{Header: m}
	case OpCodeReply:
		result = &OpReply{Header: m}
	case OpCodeGetMore:
		result = &OpGetMore{Header: m}
	case OpCodeInsert:
		result = &OpInsert{Header: m}
	case OpCodeDelete:
		result = &OpDelete{Header: m}
	case OpCodeUpdate:
		result = &OpUpdate{Header: m}
	default:
		result = &OpUnknown{Header: m}
	}
	err = result.FromReader(r)
	return result, err
}
