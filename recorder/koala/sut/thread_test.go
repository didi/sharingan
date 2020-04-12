package sut

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func Test_127_127_127_127(t *testing.T) {
	fmt.Println(binary.BigEndian.Uint32([]byte{127, 127, 127, 127}))
}
