package pthrift

import (
	"fmt"

	"github.com/modern-go/parse/model"
)

type mapVal struct {
	KeyType Kind
	ValType Kind
	Size    int
	Data    model.Map
}

func (m *mapVal) Map() model.Map {
	return model.Map{
		"key_type": m.KeyType.String(),
		"val_type": m.ValType.String(),
		"size":     m.Size,
		"data":     m.Data,
	}
}

type compactMapVal struct {
	KeyType CompactKind
	ValType CompactKind
	Size    int
	Data    model.Map
}

func (m *compactMapVal) Map() model.Map {
	return model.Map{
		"key_type": m.KeyType.String(),
		"val_type": m.ValType.String(),
		"size":     m.Size,
		"data":     m.Data,
	}
}

type listVal struct {
	ValType Kind
	Data    model.List
}

func (l *listVal) Map() model.Map {
	return model.Map{
		"val_type": l.ValType.String(),
		"data":     l.Data,
	}
}

type structVal map[int]interface{}

func (sv structVal) Map() model.Map {
	result := make(model.Map)
	for k, v := range sv {
		result[getFieldIDKey(k)] = v
	}
	return result
}

func getFieldIDKey(id int) string {
	return fmt.Sprintf("field_%d", id)
}

type mapper interface {
	Map() model.Map
}
