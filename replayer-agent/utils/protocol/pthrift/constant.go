package pthrift

import (
	"errors"
)

type MessageType byte

func (mt MessageType) String() string {
	switch mt {
	case Call:
		return "call"
	case Reply:
		return "reply"
	case Exception:
		return "exception"
	case Oneway:
		return "oneway"
	default:
		return "unknownType"
	}
}

func GetMessageType(b byte) (MessageType, error) {
	if b == 0 || b > 0x4 {
		return UnknowMessageType, errors.New("unknow message type")
	}
	return MessageType(b), nil
}

const (
	UnknowMessageType MessageType = 0x00
	Call              MessageType = 0x01
	Reply             MessageType = 0x02
	Exception         MessageType = 0x03
	Oneway            MessageType = 0x04
)

const (
	// STOP stop field
	STOP byte = 0x0
)

type Kind byte

func (k Kind) String() string {
	switch k {
	case Bool:
		return "bool"
	case Byte:
		return "byte"
	case Double:
		return "double"
	case I16:
		return "i16"
	case I32:
		return "i32"
	case I64:
		return "i64"
	case String:
		return "string"
	case Struct:
		return "struct"
	case Map:
		return "map"
	case Set:
		return "set"
	case List:
		return "list"
	default:
		return "Unknown Kind"
	}
}

const (
	UnknowKind Kind = 0x00
	Bool       Kind = 0x02
	Byte       Kind = 0x03
	Double     Kind = 0x04
	I16        Kind = 0x06
	I32        Kind = 0x08
	I64        Kind = 0x0a
	String     Kind = 0x0b
	Struct     Kind = 0x0c
	Map        Kind = 0x0d
	Set        Kind = 0x0e
	List       Kind = 0x0f
)

// GetKind ...
func GetKind(b byte) (Kind, error) {
	if b < 0x02 || b == 0x05 || b == 0x07 || b == 0x09 || b > 0x0f {
		return UnknowKind, errors.New("unknow kind")
	}
	return Kind(b), nil
}

type CompactKind byte

func (k CompactKind) String() string {
	switch k {
	case CTrue:
		return "true"
	case CFalse:
		return "false"
	case CByte:
		return "byte"
	case CDouble:
		return "double"
	case CI16:
		return "i16"
	case CI32:
		return "i32"
	case CI64:
		return "i64"
	case CBinary:
		return "string"
	case CStruct:
		return "struct"
	case CMap:
		return "map"
	case CSet:
		return "set"
	case CList:
		return "list"
	default:
		return "Unknown CompactKind"
	}
}

func (k CompactKind) ToKind() Kind {
	switch k {
	case CTrue, CFalse:
		return Bool
	case CByte:
		return Byte
	case CDouble:
		return Double
	case CI16:
		return I16
	case CI32:
		return I32
	case CI64:
		return I64
	case CBinary:
		return String
	case CStruct:
		return Struct
	case CMap:
		return Map
	case CSet:
		return Set
	case CList:
		return List
	default:
		return UnknowKind
	}
}

const (
	CUnknowKind CompactKind = 0x00
	CTrue       CompactKind = 0x01
	CFalse      CompactKind = 0x02
	CByte       CompactKind = 0x03
	CI16        CompactKind = 0x04
	CI32        CompactKind = 0x05
	CI64        CompactKind = 0x06
	CDouble     CompactKind = 0x07
	CBinary     CompactKind = 0x08
	CList       CompactKind = 0x09
	CSet        CompactKind = 0x0a
	CMap        CompactKind = 0x0b
	CStruct     CompactKind = 0x0c
)

// GetCompactKind ...
func GetCompactKind(b byte) (CompactKind, error) {
	if b == 0x0 || b > 0xc {
		return CUnknowKind, errors.New("unknow kind")
	}
	return CompactKind(b), nil
}

type ProtocolType byte

const (
	Binary ProtocolType = iota
	Compact
)
