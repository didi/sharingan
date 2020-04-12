package module

// DSLModuleData 各模块上报的dsl
var DSLModuleData map[string]map[string]int

func Init() {
	DSLModuleData = make(map[string]map[string]int)
}
