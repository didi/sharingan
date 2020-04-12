package monkey

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestMockGlobalFunc(t *testing.T) {
	type args struct {
		target      interface{}
		replacement interface{}
	}
	tests := []struct {
		name string
		args args
		want *PatchGuard
	}{
		{
			name: "test1",
			args: args{
				target: fmt.Println,
				replacement: func(a ...interface{}) (n int, err error) {
					return 0, errors.New("test error")
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MockGlobalFunc(tt.args.target, tt.args.replacement); reflect.DeepEqual(got, tt.want) {
				t.Errorf("MockGlobalFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMockMemberFunc(t *testing.T) {
	type args struct {
		target      reflect.Type
		methodName  string
		replacement interface{}
	}
	tests := []struct {
		name string
		args args
		want *PatchGuard
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				target:     reflect.TypeOf(http.DefaultClient),
				methodName: "Get",
				replacement: func(c *http.Client, url string) (*http.Response, error) {
					return nil, errors.New("test http response error")
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MockMemberFunc(tt.args.target, tt.args.methodName, tt.args.replacement); reflect.DeepEqual(got, tt.want) {
				t.Errorf("MockMemberFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}
