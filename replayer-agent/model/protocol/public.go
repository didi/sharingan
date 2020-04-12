package protocol

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

func ParsePublic(body string) (pairs map[string]json.RawMessage, requestMark string, err error) {
	splits := strings.Split(body, "||")
	if len(splits) == 0 {
		return nil, "", errors.New("malformed public log")
	}

	requestMark = splits[0]
	pairs = make(map[string]json.RawMessage)

	for _, split := range splits[1:] {
		kv := strings.Split(split, "=")
		if len(kv) != 2 {
			continue
		}
		pairs[kv[0]] = json.RawMessage(strconv.Quote(kv[1]))
	}

	return
}
