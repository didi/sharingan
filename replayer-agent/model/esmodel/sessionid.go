package esmodel

import jsoniter "github.com/json-iterator/go"

func RetrieveSessionIds(data []byte) ([]SessionId, error) {
	var source IDSource
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	err := json.Unmarshal(data, &source)
	if err != nil {
		return nil, err
	}

	var ids []SessionId
	for _, hit := range source.Hits.Hits {
		ids = append(ids, hit.IdSource)
	}

	return ids, nil
}

// ES存储的sessionID数据格式
type IDSource struct {
	Hits IDHitsOutside `json:"hits"`
}

type IDHitsOutside struct {
	Hits []IDHitsInside `json:"hits"`
}

type IDHitsInside struct {
	IdSource SessionId `json:"_source"`
}

type SessionId struct {
	Id string `json:"SessionId"`
}
