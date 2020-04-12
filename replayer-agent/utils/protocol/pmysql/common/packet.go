package common

import (
	"errors"

	"github.com/modern-go/parse"
	"github.com/modern-go/parse/read"
)

// mysql packet body length which specified by head， should larger than 0
var errPacketLengthTooSmall = errors.New("packet length too small")

// getPacketLength 获取packet的payload长度
// doc: https://dev.mysql.com/doc/internals/en/mysql-packet.html
func getPacketLength(src *parse.Source) int {
	packetLenBytes := src.ReadN(3)
	if src.FatalError() != nil {
		return 0
	}
	packetLen, _ := GetIntN(packetLenBytes, 3)
	if packetLen <= 0 {
		src.ReportError(errPacketLengthTooSmall)
		return 0
	}
	return packetLen
}

// getPacketSequenceID 获取sequenceID
// doc: https://dev.mysql.com/doc/internals/en/sequence-id.html
func getPacketSequenceID(src *parse.Source) int {
	return int(src.Read1())
}

// GetPacketHeader 获取pakcet头，简便方法，返回值第一个是payload长度，第二个是seqID
func GetPacketHeader(src *parse.Source) (int, int) {
	return getPacketLength(src), getPacketSequenceID(src)
}

// GetLenencString 获取length-encoded-string的值
// doc: https://dev.mysql.com/doc/internals/en/string.html#packet-Protocol::LengthEncodedString
func GetLenencString(src *parse.Source) string {
	n, length, err := GetIntLenc(src.PeekN(9))
	src.ResetError()
	if nil != err {
		src.ReportError(err)
		return ""
	}
	data := src.ReadN(n + length)
	if src.Error() != nil {
		return ""
	}
	str, err := GetStringFixed(data[length:], n)
	if nil != err {
		src.ReportError(err)
		return ""
	}
	return str
}

// GetLenencStringLength 根据string值判断该值编码成lenencstr要占多少字节
func GetLenencStringLength(s string) int {
	lens := len(s)
	switch {
	case lens < 251:
		return lens + 1
	case lens < (1 << 16):
		return lens + 3
	case lens < (1 << 24):
		return lens + 4
	default:
		return lens + 9
	}
}

// GetStringNull 读取string直到遇到NULL
func GetStringNull(src *parse.Source) string {
	data := read.Until1(src, 0)
	if src.Error() == nil {
		src.Read1()
	}
	return string(data)
}

// IsEOFPacket 判断是否是EOF包，使用Peek不消费字节，可重复使用
func IsEOFPacket(src *parse.Source) bool {
	data := src.PeekN(5)
	if src.Error() != nil {
		return false
	}
	length, err := GetIntN(data, 3)
	if nil != err || length == 0 {
		return false
	}
	// 要么payload一个字节，要么5个字节
	// warning: 实际抓包显示EOF包还多了两个0x00 0x00做结尾，和doc不一致
	// 暂时忽略这个长度判断
	return data[4] == 0xfe
}

// ReadEOFPacket 读取一个EOF包，如果读取成功则返回true（会消费字节）
func ReadEOFPacket(src *parse.Source) bool {
	if !IsEOFPacket(src) {
		return false
	}
	pkLen, _ := GetPacketHeader(src)
	if src.Error() != nil {
		return false
	}
	src.ReadN(pkLen)
	return src.Error() == nil
}

// ReadNULPacket 读取string[NUL]及0x00的index
// doc: https://dev.mysql.com/doc/internals/en/string.html#packet-Protocol::NulTerminatedString
func ReadNULPacket(src []byte) (bool, []byte, int) {
	posNUL := 0
	isNUL := false
	strNUL := make([]byte, 0)
	for i, b := range src {
		strNUL = append(strNUL, b)
		if b == 0x00 {
			isNUL = true
			posNUL = i
			break
		}
	}
	return isNUL, strNUL, posNUL
}
