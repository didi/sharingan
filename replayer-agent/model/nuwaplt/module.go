package nuwaplt

var Module2Host = map[string]string{}
var Host2Module = map[string]string{}
var ModuleNames = make([]Name, 0)
var ModuleNamesUniq = map[string]int{}

type Name struct {
	Name string `json:"name"`
}

func Reload() {
	//// 依据平台注册信息更新默认值
	Update()
	//for _, module := range GetModules() {
	//	Module2Host[module.Name] = module.KVs[KContext]
	//}
	//
	//for m, h := range Module2Host {
	//	Host2Module[h] = m
	//}

}

// ResetModuleInfo 清空原有数据，否则模块信息变更需要重启sharingan,影响用户体验
func ResetModuleInfo() {
	ModuleNames = make([]Name, 0)
	ModuleNamesUniq = map[string]int{}
}
