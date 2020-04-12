package pthrift

import (
	"bytes"
	"errors"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/model"
)

type messageBody struct {
	// 2 bytes version
	// ignore 1 byte
	Type       MessageType // 1 byte
	Name       string      // 4 byte length + $length bytes data
	SequenceID int         // 4 byte
}

func (mb *messageBody) Map() model.Map {
	result := make(model.Map)
	result["type"] = mb.Type.String()
	result["name"] = mb.Name
	result["sequence_id"] = mb.SequenceID
	return result
}

var (
	errNotMessage       = errors.New("not thrift message packet")
	errInvalidFeildType = errors.New("invalid field type")
)

func decodeMessage(src *parse.Source) (*messageBody, error) {
	// version 0x8001
	if !src.Expect1(byte(0x80)) {
		return nil, errNotMessage
	}
	if !src.Expect1(byte(0x01)) {
		return nil, errNotMessage
	}
	// ignore unused 1 byte
	src.Read1()
	if src.Error() != nil {
		return nil, src.Error()
	}
	body := &messageBody{}
	// type
	t, err := GetMessageType(src.Read1())
	if nil != err {
		return nil, err
	}
	if src.Error() != nil {
		return nil, src.Error()
	}
	body.Type = t

	// 4 bytes length + $length bytes data
	length, err := getInt(src, 4)
	if nil != err {
		return nil, err
	}
	name := src.ReadN(length)
	if src.Error() != nil {
		return nil, src.Error()
	}
	body.Name = string(name)

	// 4 bytes sequence id
	sequenceID, err := getInt(src, 4)
	if nil != err {
		return nil, err
	}
	body.SequenceID = sequenceID
	return body, nil
}

func decodeMap(src *parse.Source) (*mapVal, error) {
	keyType, err := GetKind(src.Read1())
	if nil != err {
		return nil, err
	}
	valType, err := GetKind(src.Read1())
	if nil != err {
		return nil, err
	}
	size, err := getInt(src, 4)
	if nil != err {
		return nil, err
	}
	result := &mapVal{
		KeyType: keyType,
		ValType: valType,
		Size:    size,
		Data:    make(model.Map),
	}
	for i := 0; i < size; i++ {
		k, err := decodeFieldValue(src, keyType)
		if nil != err {
			return nil, err
		}
		v, err := decodeFieldValue(src, valType)
		if nil != err {
			return nil, err
		}
		result.Data[k] = v
	}
	return result, src.Error()
}

func decodeList(src *parse.Source) (*listVal, error) {
	valType, err := GetKind(src.Read1())
	if nil != err {
		return nil, err
	}
	size, err := getInt(src, 4)
	if nil != err {
		return nil, err
	}
	result := &listVal{
		ValType: valType,
	}
	for i := 0; i < size; i++ {
		val, err := decodeFieldValue(src, valType)
		if nil != err {
			return nil, err
		}
		result.Data = append(result.Data, val)
	}
	return result, src.Error()
}

func decodeStruct(src *parse.Source) (structVal, error) {
	result := make(structVal)
	for src.Error() == nil && !src.Expect1(STOP) {
		vType, err := GetKind(src.Read1())
		if nil != err {
			return nil, err
		}
		fieldID, err := getInt(src, 2)
		if nil != err {
			return nil, err
		}
		v, err := decodeFieldValue(src, vType)
		if nil != err {
			return nil, err
		}
		result[fieldID] = v
	}
	return result, src.Error()
}

func decodeFieldValue(src *parse.Source, vType Kind) (out interface{}, err error) {
	switch vType {
	case Bool:
		return getBool(src.Read1()), src.Error()
	case Byte:
		return src.Read1(), src.Error()
	case Double:
		return getDouble(src)
	case I16:
		return getInt(src, 2)
	case I32:
		return getInt(src, 4)
	case I64:
		return getInt(src, 8)
	case String:
		return getString(src)
	case Struct:
		v, err := decodeStruct(src)
		if nil != err {
			return nil, err
		}
		return v.Map(), err
	case Map:
		v, err := decodeMap(src)
		if nil != err {
			return nil, err
		}
		return v.Map(), err
	case Set, List:
		v, err := decodeList(src)
		if nil != err {
			return nil, err
		}
		return v.Map(), err
	default:
		return nil, errInvalidFeildType
	}
}

// DecodeBinary 解析thrift binary协议
func DecodeBinary(b []byte) (model.Map, error) {
	return decodeThrift(b, Binary)
}

// DecodeCompact 解析thrift compact协议
func DecodeCompact(b []byte) (model.Map, error) {
	return decodeThrift(b, Compact)
}

func decodeThrift(b []byte, t ProtocolType) (model.Map, error) {
	src, err := parse.NewSource(bytes.NewBuffer(b), 20)
	if nil != err {
		return nil, err
	}
	var messageDecoder func(*parse.Source) (*messageBody, error)
	var structDecoder func(*parse.Source) (structVal, error)
	if t == Binary {
		messageDecoder = decodeMessage
		structDecoder = decodeStruct
	} else {
		messageDecoder = decodeMessageCompact
		structDecoder = decodeStructCompact
	}
	messageBody, err := messageDecoder(src)
	if nil != err {
		return nil, err
	}
	val, err := structDecoder(src)
	if nil != err {
		return nil, err
	}
	result := messageBody.Map()
	result["param"] = val.Map()
	return result, nil
}

func decodeMessageCompact(src *parse.Source) (*messageBody, error) {
	if !src.Expect1(0x82) {
		return nil, errNotMessage
	}
	b := src.Read1()
	mType, err := GetMessageType(b >> 5)
	if nil != err {
		return nil, err
	}
	seqID, err := readVarint32(src)
	if nil != err {
		return nil, err
	}
	name, err := readVarString(src)
	if nil != err {
		return nil, err
	}
	return &messageBody{
		Type:       mType,
		Name:       name,
		SequenceID: int(seqID),
	}, nil
}

func decodeListCompact(src *parse.Source) (*listVal, error) {
	b := src.Read1()
	if err := src.Error(); err != nil {
		return nil, err
	}
	valType, err := GetCompactKind(b & 0xf)
	if nil != err {
		return nil, err
	}
	b >>= 4
	size := int32(0)
	if b < 0xf {
		size = int32(b)
	} else {
		size, err = readVarint32(src)
		if nil != err {
			return nil, err
		}
		if size < 15 {
			return nil, errNotMessage
		}
	}
	list := make(model.List, size)
	for i := int32(0); i < size; i++ {
		list[i], err = decodeFieldValueCompact(src, valType)
		if nil != err {
			return nil, err
		}
	}
	return &listVal{
		ValType: valType.ToKind(),
		Data:    list,
	}, nil
}

func decodeMapCompact(src *parse.Source) (*mapVal, error) {
	// empty map
	if src.Expect1(0) {
		return nil, nil
	}

	//non-empty map
	size, err := readVarint32(src)
	if nil != err {
		return nil, err
	}
	b := src.Read1()
	if err := src.Error(); err != nil {
		return nil, err
	}
	keyType, err := GetCompactKind(b >> 4)
	if nil != err {
		return nil, err
	}
	vType, err := GetCompactKind(b & 0xf)
	if nil != err {
		return nil, err
	}
	result := &mapVal{
		KeyType: keyType.ToKind(),
		ValType: vType.ToKind(),
		Size:    int(size),
		Data:    make(model.Map),
	}
	for i := int32(0); i < size; i++ {
		key, err := decodeFieldValueCompact(src, keyType)
		if nil != err {
			return nil, err
		}
		var val interface{}
		if keyType.ToKind() != Bool {
			val, err = decodeFieldValueCompact(src, vType)
			if nil != err {
				return nil, err
			}
		} else {
			if keyType == CTrue {
				val = true
			} else {
				val = false
			}
		}
		result.Data[key] = val
	}
	return result, nil
}

func decodeStructCompact(src *parse.Source) (structVal, error) {
	previousFieldID := int32(0)
	var err error
	result := make(structVal)
	for src.Error() == nil && !src.Expect1(STOP) {
		b := src.Read1()
		// 0000tttt
		fieldID := int32(0)
		delta := (b & 0xf0) >> 4
		if delta == 0 {
			fieldID, err = readVarint32(src)
			if nil != err {
				return nil, err
			}
		} else {
			// ddddtttt
			fieldID = previousFieldID + int32(delta)
		}
		previousFieldID = fieldID
		vType, err := GetCompactKind(b & 0xf)
		if nil != err {
			return nil, err
		}
		val, err := decodeFieldValueCompact(src, vType)
		if nil != err {
			return nil, err
		}
		if val != nil {
			result[int(fieldID)] = val
		}
	}
	return result, nil
}

func decodeFieldValueCompact(src *parse.Source, vType CompactKind) (interface{}, error) {
	switch vType {
	case CTrue:
		return true, nil
	case CFalse:
		return false, nil
	case CDouble:
		return getCompactDouble(src)
	case CI16, CI32:
		v, err := readVarint32(src)
		return int(zigzag2I32(v)), err
	case CI64:
		v, err := readVarint64(src)
		return int(zigzag2I64(v)), err
	case CBinary:
		return readVarString(src)
	case CStruct:
		v, err := decodeStructCompact(src)
		if nil != err {
			return nil, err
		}
		return v.Map(), err
	case CMap:
		v, err := decodeMapCompact(src)
		if nil != err {
			return nil, err
		}
		return v.Map(), err
	case CSet, CList:
		v, err := decodeListCompact(src)
		if nil != err {
			return nil, err
		}
		return v.Map(), err
	default:
		return nil, errInvalidFeildType
	}
}
