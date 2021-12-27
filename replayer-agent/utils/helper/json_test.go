package helper

import (
	"fmt"
	"testing"
)

func TestJson2SingleLayerMap(t *testing.T) {

	str := `
		{"int": 1, 
		"str":"string", 
		"map":{"str":"string", "float":12.22},
		"slice":[1,2,3,"4"]}`
	m, _ := Json2SingleLayerMap([]byte(str))

	// int 1
	// str "string"
	// map.str "string"
	// map.float 12.22
	// slice.2 3
	// slice.3 "4"
	// slice.0 1
	// slice.1 2
	for k, v := range m {
		fmt.Println(k, string(v))
	}

	raw, ok := m["slice.0"]
	if !ok || string(raw) != "1" {
		t.Error()
	}
}

func TestWrappedJsonString(t *testing.T) {

	str := `{"a":"{\"b\":\"c\"}"}`
	m, _ := Json2SingleLayerMap([]byte(str))

	for k, v := range m {
		fmt.Println(k, string(v))
	}
}
