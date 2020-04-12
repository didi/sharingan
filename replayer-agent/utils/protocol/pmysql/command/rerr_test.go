package command

import (
	"bytes"
	"testing"

	"github.com/modern-go/parse"
	"github.com/stretchr/testify/require"
)

func TestDecodeErrPacket(t *testing.T) {
	var testCase = []struct {
		rawBytes []byte
		expect   *ErrResp
		err      error
	}{
		{
			rawBytes: []byte{
				0x2a, 0x00, 0x00, 0x01, 0xff, 0x7a, 0x04, 0x23, 0x34, 0x32, 0x53, 0x30,
				0x32, 0x54, 0x61, 0x62, 0x6c, 0x65, 0x20, 0x27, 0x74, 0x65, 0x73, 0x74,
				0x2e, 0x73, 0x61, 0x6c, 0x61, 0x72, 0x79, 0x27, 0x20, 0x64, 0x6f, 0x65,
				0x73, 0x6e, 0x27, 0x74, 0x20, 0x65, 0x78, 0x69, 0x73, 0x74,
			},
			expect: &ErrResp{
				Header:     0xff,
				ErrCode:    1146,
				ExtraBytes: []byte("#42S02Table 'test.salary' doesn't exist"),
			},
			err: nil,
		},
	}
	should := require.New(t)
	for idx, tc := range testCase {
		src, err := parse.NewSource(bytes.NewReader(tc.rawBytes), 30)
		should.NoError(err)
		actual, err := DecodeErrPacket(src)
		should.Equal(tc.err, err, "case #%d fail", idx)
		should.Equal(tc.expect.String(), actual.String(), "case #%d fail", idx)
	}
}
