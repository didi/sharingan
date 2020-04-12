package pthrift

import (
	"bytes"
	"testing"

	"github.com/modern-go/parse"
	"github.com/stretchr/testify/require"
)

func TestGetDouble(t *testing.T) {
	var testCase = []struct {
		b      []byte
		expect float64
		err    error
	}{
		{
			b:      []byte{0x40, 0x59, 0x38, 0x00, 0x00, 0x00, 0x00, 0x00},
			expect: 100.875,
		},
	}
	should := require.New(t)
	for _, tc := range testCase {
		src, err := parse.NewSource(bytes.NewBuffer(tc.b), 3)
		should.NoError(err)
		actual, err := getDouble(src)
		should.Equal(tc.err, err)
		should.Equal(tc.expect, actual, "%x", tc.b)
	}
}
