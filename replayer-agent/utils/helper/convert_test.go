package helper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStripSlashes(t *testing.T) {
	should := require.New(t)
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			"t1",
			args{data: []byte("How\\'s everybody")},
			[]byte("How's everybody"),
		},
		{
			"t2",
			args{data: []byte("Are you \\\"JOHN\\\"?")},
			[]byte("Are you \"JOHN\"?"),
		},
		{
			"t3",
			args{data: []byte("c:\\\\php\\\\stripslashes")},
			[]byte("c:\\php\\stripslashes"),
		},
		{
			"t4",
			args{data: []byte("c:\\php\\stripslashes")},
			[]byte("c:phpstripslashes"),
		},
		{
			"t5",
			args{data: []byte("\\xff")},
			[]byte{255},
		},
	}
	for _, tt := range tests {
		got := StripcSlashes(tt.args.data)
		should.Equal(tt.want, got,
			fmt.Sprintf("%q. StripcSlashes() want %v, = %v", tt.name, tt.want, got))
	}
}

func Test_isxdigit(t *testing.T) {
	should := require.New(t)
	type args struct {
		b byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"t1",
			args{'3'},
			true,
		},
		{
			"t2",
			args{'\a'},
			false,
		},
	}
	for _, tt := range tests {
		got := isxdigit(tt.args.b)
		should.Equal(tt.want, got,
			fmt.Sprintf("%q. isxdigit() want %v, = %v", tt.name, tt.want, got))
	}
}

func Test_bytetol(t *testing.T) {
	should := require.New(t)
	type args struct {
		buf  []byte
		base int
	}
	tests := []struct {
		name string
		args args
		want byte
	}{
		{
			"t1",
			args{[]byte{'f', 'f'}, 16},
			byte(255),
		},
		{
			"t2",
			args{[]byte{'0', '4'}, 8},
			byte(4),
		},
	}
	for _, tt := range tests {
		got := strtol(tt.args.buf, tt.args.base)
		should.Equal(tt.want, got,
			fmt.Sprintf("%q. bytetol() want %v, = %v", tt.name, tt.want, got))
	}
}
