package helper

import (
	"encoding/json"
	"strconv"
)

// json解析成单层map, 打平多层嵌套,多层嵌套通过"."连接key
// 对于数组采用数字作为key
func Json2SingleLayerMap(body []byte) (map[string]json.RawMessage, error) {
	singleLayerMap := make(map[string]json.RawMessage)
	var (
		msgMap   map[string]json.RawMessage
		msgSlice []json.RawMessage
	)

	if err := json.Unmarshal(body, &msgMap); err == nil {
		for key, value := range msgMap {
			key = BytesToString(Decode(StringToBytes(key)))
			smap, err := Json2SingleLayerMap(value)
			if err != nil || len(smap) == 0 {
				singleLayerMap[key] = value
			} else {
				for k, v := range smap {
					smapKey := key + "." + k
					singleLayerMap[smapKey] = v
				}
			}
		}
	} else {
		err = json.Unmarshal(body, &msgSlice)
		if err != nil {
			return singleLayerMap, err
		}
		for key, value := range msgSlice {
			smap, err := Json2SingleLayerMap(value)
			if err != nil {
				smapKey := strconv.Itoa(key)
				singleLayerMap[smapKey] = value
			} else {
				for k, v := range smap {
					smapKey := strconv.Itoa(key) + "." + k
					singleLayerMap[smapKey] = v
				}
			}
		}
	}

	return singleLayerMap, nil
}
