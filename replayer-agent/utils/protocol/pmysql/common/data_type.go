package common

import (
	"errors"

	"github.com/modern-go/parse"
)

const (
	// NULL 结果集中如果某个字段是NULL，用0xfe表示
	NULL = byte(0xfb)
)

// GetIntN 获取fixedInteger int<N>的值
// doc: https://dev.mysql.com/doc/internals/en/integer.html#fixed-length-integer
func GetIntN(data []byte, n int) (int, error) {
	if len(data) < n {
		return -1, errPacketLengthTooSmall
	}
	val := uint32(data[0])
	pos := uint(8)
	for i := 1; i < n; i++ {
		val |= uint32(data[i]) << pos
		pos += 8
	}
	return int(val), nil
}

// GetIntLenc 获取Length-Encoded-Integer的值
// 第二个返回值表示编码int相关的所有字节的长度
// doc: https://dev.mysql.com/doc/internals/en/integer.html#fixed-length-integer
func GetIntLenc(data []byte) (int, int, error) {
	if 0 == len(data) {
		return -1, 0, errPacketLengthTooSmall
	}
	v := data[0]
	if v < 251 {
		return int(v), 1, nil
	}
	switch v {
	case 0xfc:
		val, err := GetIntN(data[1:], 2)
		return val, 3, err
	case 0xfd:
		val, err := GetIntN(data[1:], 3)
		return val, 4, err
	case 0xfe:
		v, err := GetIntN(data[1:], 8)
		if err == errPacketLengthTooSmall {
			return -1, 0, errors.New("try to read integer from EOF packet")
		}
		return v, 9, err
	default:
		return -1, 0, errors.New("first byte of lenEcodedInt is 0xff")
	}
}

// GetLenencInt 获取Length-Encoded-Integer的值
// 返回值第二个int表示编码该int总共的字节数
func GetLenencInt(src *parse.Source) (int, int, error) {
	v := src.Read1()
	if src.Error() != nil {
		return 0, 1, src.Error()
	}
	if v < 251 {
		return int(v), 1, nil
	}
	switch v {
	case 0xfc:
		data := append([]byte{v}, src.ReadN(2)...)
		return GetIntLenc(data)
	case 0xfd:
		data := append([]byte{v}, src.ReadN(3)...)
		return GetIntLenc(data)
	case 0xfe:
		data := append([]byte{v}, src.ReadN(8)...)
		return GetIntLenc(data)
	default:
		return -1, 0, errors.New("first byte of lenEcodedInt is 0xff")
	}
}

// GetStringFixed 读取固定长度的string
func GetStringFixed(data []byte, n int) (string, error) {
	if len(data) < n {
		return "", errPacketLengthTooSmall
	}
	return string(data[:n]), nil
}
