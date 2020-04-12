package match

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_findReadableChunk(t *testing.T) {
	should := require.New(t)
	type args struct {
		key []byte
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 int
	}{
		{
			"t1",
			args{[]byte("*2\r\n$6\r\nEXISTS\r\n$19\r\nDB_CLIENT_INFO_amap\r\n")},
			0,
			2,
		},
		{
			"t2",
			args{[]byte("\r\n$6\r\nEXISTS\r\n$19\r\nDB_CLIENT_INFO_amap\r\n")},
			2,
			2,
		},
	}
	for _, tt := range tests {
		got, got1 := findReadableChunk(tt.args.key)

		should.Equal(tt.want, got,
			fmt.Sprintf("%q. findReadableChunk() want %v, got = %v", tt.name, tt.want, got))

		should.Equal(tt.want1, got1,
			fmt.Sprintf("%q. findReadableChunk() want %v, got1 = %v", tt.name, tt.want1, got1))
	}
}

func Test_cutToChunks(t *testing.T) {
	type args struct {
		key  []byte
		unit int
	}
	tests := []struct {
		name string
		args args
		want [][]byte
	}{
		{"1",
			args{[]byte("GET /passport/login/v5/refreshTicket?access_key_id=1&maptype=soso&terminal_id=1 HTTP/ asdas"), 16},
			[][]byte{
				[]byte("GET /passport/login/v5/refreshTicket"),
				[]byte("access_key_id=1"),
				[]byte("maptype=soso"),
				[]byte("terminal_id=1"),
				[]byte("HTTP/ asdas"),
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cutToChunks(tt.args.key, tt.args.unit); !reflect.DeepEqual(got, tt.want) {
				for _, g := range got {
					t.Errorf("cutToChunks() = %v", string(g))
				}
				t.Errorf("want %v", tt.want)
			}
		})
	}
}
