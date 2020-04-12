package monkey

import (
	"reflect"

	"github.com/agiledragon/gomonkey"
)

// PatchGuard PatchGuard
type PatchGuard struct {
	*gomonkey.Patches
}

// MockGlobalFunc MockGlobalFunc
func MockGlobalFunc(target, replacement interface{}) *PatchGuard {
	return &PatchGuard{gomonkey.ApplyFunc(target, replacement)}
}

// MockMemberFunc MockMemberFunc
func MockMemberFunc(target reflect.Type, methodName string, replacement interface{}) *PatchGuard {
	return &PatchGuard{gomonkey.ApplyMethod(target, methodName, replacement)}
}
