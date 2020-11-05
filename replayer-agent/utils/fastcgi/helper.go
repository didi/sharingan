package fastcgi

import (
	"bytes"
	"encoding/binary"
	"io"
)

/**
 * helper functions
 */

func (h *Header) ReadLength(r io.Reader, order binary.ByteOrder, data *int) error {
	var b [8]byte
	bs := b[:1]
	if _, err := io.ReadFull(r, bs); err != nil {
		return err
	}
	if bs[0]&0x80 != 0 {
		bs[0] ^= 0x80
		bs = b[1:4]
		if _, err := io.ReadFull(r, bs); err != nil {
			return err
		}
		bs = b[:4]
	}

	if len(bs) == 1 {
		*data = int(bs[0])
	} else {
		*data = int(order.Uint32(bs))
	}
	h.ContentLength -= uint16(len(bs) + (*data))
	return nil
}

func (h *Header) skipPaddings(buf *bytes.Buffer) {
	if h.PaddingLength > 0 {
		data := make([]byte, h.PaddingLength)
		binary.Read(buf, binary.BigEndian, data)
	}
}

func merge(old map[string]string, new map[string]string) map[string]string {
	if old == nil || len(old) == 0 {
		return new
	}
	for k, v := range new {
		old[k] = v
	}
	return old
}
