package recorder

import "testing"

func TestSetDelegatedFromGoRoutineID(t *testing.T) {
	type args struct {
		gID int64
	}
	tests := []struct {
		name string
		args args
	}{
		{"1", args{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDelegatedFromGoRoutineID(tt.args.gID)
			goid := GetCurrentGoRoutineID()
			if goid != 0 {
				t.Errorf("GetCurrentGoRoutineID() = %v, want %v", goid, 0)
			}
		})
	}
}
