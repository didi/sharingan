/**
 * library to parse fastcgi protocol
 */
package fastcgi

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	HeaderSize = 8
)

const (
	Version = uint8(1)
)

// recType is a record type, as defined by
// https://web.archive.org/web/20150420080736/http://www.fastcgi.com/drupal/node/6?q=node/22#S8
type recType uint8

const (
	TypeBeginRequest    recType = 1
	TypeAbortRequest    recType = 2
	TypeEndRequest      recType = 3
	TypeParams          recType = 4
	TypeStdin           recType = 5
	TypeStdout          recType = 6
	TypeStderr          recType = 7
	TypeData            recType = 8
	TypeGetValues       recType = 9
	TypeGetValuesResult recType = 10
	TypeUnknownType     recType = 11
)

const (
	statusRequestComplete = iota
	statusCantMultiplex
	statusOverloaded
	statusUnknownRole
)

const (
	roleResponder = iota + 1 // only Responders are implemented.
	roleAuthorizer
	roleFilter
)

type Header struct {
	Version       uint8
	Type          recType
	Id            uint16
	ContentLength uint16
	PaddingLength uint8
	Reserved      uint8
}

func ParseHeader(buf *bytes.Buffer) (*Header, error) {
	if buf == nil {
		return nil, errors.New("empty buffer")
	}
	var fh Header
	err := binary.Read(buf, binary.BigEndian, &fh)
	if err != nil {
		return nil, err
	}
	// 类型校验
	if fh.Version != 1 && fh.Type > TypeUnknownType {
		return nil, errors.New("unknown message type")
	}
	if buf.Len() < int(fh.ContentLength)+int(fh.PaddingLength) {
		return nil, errors.New("truncated buffer")
	}
	return &fh, nil
}

type Record interface {
	GetType() recType
}

func (h *Header) ParseRecord(buf *bytes.Buffer) (Record, error) {
	switch h.Type {
	case TypeBeginRequest:
		return h.parseBeginRequest(buf)
	case TypeAbortRequest:
	case TypeEndRequest:
		return h.parseEndRequest(buf)
	case TypeParams:
		return h.parseParams(buf)
	case TypeStdin:
		return h.parseStdin(buf)
	case TypeStdout:
		return h.parseStdout(buf)
	case TypeStderr:
		return h.parseStderr(buf)
	case TypeData:
	case TypeGetValues:
	case TypeGetValuesResult:
	case TypeUnknownType:
		return h.parseUnknown(buf)
	}
	return nil, nil
}

type BeginRequest struct {
	Role     uint16
	Flags    uint8
	Reserved [5]uint8
}

func (br *BeginRequest) GetType() recType {
	return TypeBeginRequest
}

func (h *Header) parseBeginRequest(buf *bytes.Buffer) (*BeginRequest, error) {
	var beginRequest BeginRequest
	err := binary.Read(buf, binary.BigEndian, &beginRequest)
	if err != nil {
		return nil, err
	}
	h.skipPaddings(buf)
	return &beginRequest, err
}

type EndRequest struct {
	AppStatus      uint32
	ProtocolStatus uint8
	Reserved       [3]uint8
}

func (br *EndRequest) GetType() recType {
	return TypeEndRequest
}

func (h *Header) parseEndRequest(buf *bytes.Buffer) (*EndRequest, error) {
	var endRequest EndRequest
	err := binary.Read(buf, binary.BigEndian, &endRequest)
	if err != nil {
		return nil, err
	}
	h.skipPaddings(buf)
	return &endRequest, nil
}

type Params struct {
	Maps map[string]string
}

func (br *Params) GetType() recType {
	return TypeParams
}

func (h *Header) parseParams(buf *bytes.Buffer) (*Params, error) {
	params := Params{Maps: make(map[string]string)}
	for h.ContentLength > 0 {
		var nameLen, valueLen int
		err := h.ReadLength(buf, binary.BigEndian, &nameLen)
		if err != nil {
			return nil, err
		}
		err = h.ReadLength(buf, binary.BigEndian, &valueLen)
		if err != nil {
			return nil, err
		}

		nb := make([]byte, nameLen)
		err = binary.Read(buf, binary.BigEndian, nb)
		if err != nil {
			return nil, err
		}
		vb := make([]byte, valueLen)
		err = binary.Read(buf, binary.BigEndian, vb)
		if err != nil {
			return nil, err
		}

		params.Maps[string(nb)] = string(vb)
	}
	h.skipPaddings(buf)
	return &params, nil
}

type StdinRequest struct {
	Data []byte
}

func (br *StdinRequest) GetType() recType {
	return TypeStdin
}

func (h *Header) parseStdin(buf *bytes.Buffer) (*StdinRequest, error) {
	res := make([]byte, h.ContentLength)
	err := binary.Read(buf, binary.BigEndian, res)
	if err != nil {
		return nil, err
	}
	h.skipPaddings(buf)
	return &StdinRequest{Data: res}, nil
}

type StdoutResponse struct {
	Data []byte
}

func (br *StdoutResponse) GetType() recType {
	return TypeStdout
}

func (h *Header) parseStdout(buf *bytes.Buffer) (*StdoutResponse, error) {
	data := make([]byte, h.ContentLength)
	err := binary.Read(buf, binary.BigEndian, data)
	if err != nil {
		return nil, err
	}

	h.skipPaddings(buf)

	return &StdoutResponse{Data: data}, nil
}

type StderrResponse struct {
	Data []byte
}

func (br *StderrResponse) GetType() recType {
	return TypeStderr
}

func (h *Header) parseStderr(buf *bytes.Buffer) (*StderrResponse, error) {
	data := make([]byte, h.ContentLength)
	err := binary.Read(buf, binary.BigEndian, data)
	if err != nil {
		return nil, err
	}

	h.skipPaddings(buf)

	return &StderrResponse{Data: data}, nil
}

type UnknownType struct {
	Type     uint8
	Reserved [7]uint8
}

func (br *UnknownType) GetType() recType {
	return TypeUnknownType
}

func (h *Header) parseUnknown(buf *bytes.Buffer) (*UnknownType, error) {
	var unknown UnknownType
	err := binary.Read(buf, binary.BigEndian, &unknown)
	if err != nil {
		return nil, err
	}
	h.skipPaddings(buf)
	return &unknown, nil
}
