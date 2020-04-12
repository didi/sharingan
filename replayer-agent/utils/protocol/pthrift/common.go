package pthrift

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/modern-go/parse"
)

var (
	errNotI16    = errors.New("not i16 type")
	errNotI32    = errors.New("not i32 type")
	errNotI64    = errors.New("not i64 type")
	errNotStr    = errors.New("not string type")
	errNotDouble = errors.New("not double type")
	errNotMap    = errors.New("not map type")
)

func getInt(src *parse.Source, length int) (int, error) {
	b := src.ReadN(length)
	switch length {
	case 2:
		return getInt16(b)
	case 4:
		return getInt32(b)
	case 8:
		return getInt64(b)
	default:
		return 0, errors.New("invalid int length")
	}
}

func getInt16(b []byte) (int, error) {
	if len(b) < 2 {
		return 0, errNotI16
	}
	return int(binary.BigEndian.Uint16(b)), nil
}

func getInt32(b []byte) (int, error) {
	if len(b) < 4 {
		return 0, errNotI32
	}
	return int(binary.BigEndian.Uint32(b)), nil
}

func getInt64(b []byte) (int, error) {
	if len(b) < 8 {
		return 0, errNotI64
	}
	return int(binary.BigEndian.Uint64(b)), nil
}

func getBool(b byte) bool {
	return b == 0x01
}

func getString(src *parse.Source) (string, error) {
	length, err := getInt(src, 4)
	if nil != err {
		return "", err
	}
	if length < 0 {
		return "", errNotStr
	}
	bytes := src.ReadN(length)
	return string(bytes), src.Error()
}

func getDouble(src *parse.Source) (float64, error) {
	b := src.ReadN(8)
	if len(b) < 8 {
		return 0, errNotDouble
	}
	return math.Float64frombits(binary.BigEndian.Uint64(b)), src.Error()
}

func getCompactDouble(src *parse.Source) (float64, error) {
	b := src.ReadN(8)
	if len(b) < 8 {
		return 0, errNotDouble
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(b)), src.Error()
}

func zigzag2I32(n int32) int32 {
	u := uint32(n)
	return int32(u>>1) ^ -(n & 1)
}

func zigzag2I64(n int64) int64 {
	u := uint64(n)
	return int64(u>>1) ^ -(n & 1)
}

func readVarint32(src *parse.Source) (int32, error) {
	ui64, err := binary.ReadUvarint(src)
	if nil != err {
		return 0, err
	}
	return int32(ui64), nil
	/*
		i64, err := readVarint64FromWire(src)
		if err != nil {
			return 0, err
		}
		return zigzag2I32(int32(i64)), nil
	*/
}

func readVarint64(src *parse.Source) (int64, error) {
	ui64, err := binary.ReadUvarint(src)
	if nil != err {
		return 0, err
	}
	return int64(ui64), nil
	/*
		i64, err := readVarint64FromWire(src)
		if err != nil {
			return 0, err
		}
		return zigzag2I64(i64), nil
	*/
}

func readVarint64FromWire(src *parse.Source) (int64, error) {
	shift := uint(0)
	result := int64(0)
	for {
		b := src.Read1()
		if src.Error() != nil {
			return 0, src.Error()
		}
		result |= int64(b&0x7f) << shift
		if (b & 0x80) != 0x80 {
			break
		}
		shift += 7
	}
	return result, nil
}

func readVarString(src *parse.Source) (string, error) {
	length, err := readVarint32(src)
	if nil != err {
		return "", err
	}
	return string(src.ReadN(int(length))), src.Error()
}
