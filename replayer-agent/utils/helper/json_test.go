package helper

import (
	"testing"
)

func TestJson2SingleLayerMap(t *testing.T) {
	str := `
		{"int": 1, 
		"str":"string", 
		"map":{"str":"string", "float":12.22},
		"slice":[1,2,3,"4"]}`
	m, _ := Json2SingleLayerMap([]byte(str))
	raw, ok := m["slice.0"]
	if !ok || string(raw) != "1" {
		t.Error()
	}
}
