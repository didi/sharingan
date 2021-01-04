package hook

import (
	"net"
	"testing"
)

func Test_ip6ZoneToString(t *testing.T) {
	type args struct {
		zone int
	}

	nets, _ := net.Interfaces()
	want := nets[0].Name

	tests := []struct {
		name string
		args args
		want string
	}{
		{"1", args{1}, want},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ip6ZoneToString(tt.args.zone); got != tt.want {
				t.Errorf("ip6ZoneToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_itod(t *testing.T) {
	type args struct {
		i uint
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1", args{0}, "0"},
		{"1", args{12}, "12"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := itod(tt.args.i); got != tt.want {
				t.Errorf("itod() = %v, want %v", got, tt.want)
			}
		})
	}
}
