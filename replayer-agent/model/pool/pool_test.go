package pool

import "testing"

func TestGetBuf(t *testing.T) {
	t.Run("nozero", func(t *testing.T) {
		cases := map[string]struct {
			Size   int
			Zero   bool
			ExpLen int
			ExpCap int
		}{
			"nz1": {1, false, 32, 32},
			"nz2": {17, false, 32, 32},
			"nz3": {139307, false, 163840, 163840},
		}
		for name, cas := range cases {
			buf := GetBuf(cas.Size, cas.Zero)
			if len(buf) != cas.ExpLen || cap(buf) != cas.ExpCap {
				t.Fatalf("case: %s  expect: len=%d, cap=%d; got: len=%d, cap=%d", name, cas.ExpLen, cas.ExpCap, len(buf), cap(buf))
			}
			PutBuf(buf)
		}
	})
	t.Run("zero", func(t *testing.T) {
		cases := map[string]struct {
			Size   int
			Zero   bool
			ExpLen int
			ExpCap int
		}{
			"nz1": {1, true, 0, 32},
			"nz2": {17, true, 0, 32},
			"nz3": {139307, true, 0, 163840},
		}
		for name, cas := range cases {
			buf := GetBuf(cas.Size, cas.Zero)
			if len(buf) != cas.ExpLen || cap(buf) != cas.ExpCap {
				t.Fatalf("case: %s  expect: len=%d, cap=%d; got: len=%d, cap=%d", name, cas.ExpLen, cas.ExpCap, len(buf), cap(buf))
			}
			PutBuf(buf)
		}
	})
}
