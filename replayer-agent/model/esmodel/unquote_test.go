package esmodel

import "testing"

func TestUnquote(t *testing.T) {
	type args struct {
		s []byte
	}
	tests := []struct {
		name  string
		args  args
		wantT string
	}{
		{"1", args{s: []byte(`"\uD925\uDFA1"`)}, "񙞡"},
		{"2", args{s: []byte(`"\u4e16\u754c"`)}, "世界"},
		{"3", args{s: []byte(`"abc123"`)}, "abc123"},
		{"4", args{s: []byte(`""`)}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotT, _ := Unquote(tt.args.s)
			if gotT != tt.wantT {
				t.Errorf("Unquote() gotT = %v, want %v", gotT, tt.wantT)
			}
		})
	}
}
