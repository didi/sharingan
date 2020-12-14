package nuwaplt

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"github.com/didi/sharingan/replayer-agent/common/handlers/conf"
	"github.com/didi/sharingan/replayer-agent/common/handlers/httpclient"
	"github.com/didi/sharingan/replayer-agent/common/handlers/tlog"
	"github.com/didi/sharingan/replayer-agent/utils/helper"
)

const (
	PHttp             = "http"
	LGO               = "go"
	DefaultDepartment = "default"

	// key
	KContext    = "context"
	KListenAddr = "listen-addr"
	KProtocol   = "protocol"
	KLanguage   = "language"
	KDepartment = "department"
	KLogPath    = "logpath"
	KServerType = "server-type"
)

type Modules struct {
	sync.RWMutex `json:"-"`

	Data []*Module `json:"data"`
}
type Module struct {
	Name string            `json:"name"`
	Data Maps              `json:"data"`
	KVs  map[string]string `json:"-"`
}
type Maps struct {
	M []map[string]string
}

var AllModules Modules

func Update() {
	AllModules.Update()
}

func GetModules() []*Module {
	AllModules.RLock()
	defer AllModules.RUnlock()

	return AllModules.Data
}

func GetValueWithProject(sProject, sKey, dValDefault string) string {
	AllModules.RLock()
	defer AllModules.RUnlock()

	for _, module := range AllModules.Data {
		if module.Name != sProject {
			continue
		}
		val, ok := module.KVs[sKey]
		if ok {
			return val
		}
	}
	return dValDefault
}

func GetValueByKey(project, dKey, dValDefault string) string {
	AllModules.RLock()
	defer AllModules.RUnlock()

	for _, module := range AllModules.Data {
		if module.Name == project {
			dVal, ok := module.KVs[dKey]
			if ok {
				return dVal
			}
		}
	}
	return dValDefault
}

func (m *Modules) Update() {
	m.Lock()
	defer m.Unlock()

	var res []byte
	var err error
	url := conf.Handler.GetString("http_api.module_info")
	if url != "" {
		_, res, err = httpclient.Handler.Get(context.Background(), url)
		if err != nil {
			tlog.Handler.Errorf(context.Background(), tlog.DLTagUndefined, "query module infomation from nuwa platform failed||err=%s", err)
			return
		}
	} else {
		//读取配置文件conf/moduleinfo.json
		res, err = helper.ReadFileBytes(conf.Root + "/conf/moduleinfo.json")
		if err != nil {
			tlog.Handler.Errorf(context.Background(), tlog.DLTagUndefined, "Failed to read /conf/moduleinfo.json err="+err.Error())
			return
		}
	}

	json.Unmarshal(res, m)
	// 重置模块信息
	ResetModuleInfo()
	for _, module := range m.Data {
		module.Fulfill()
	}
}

func (m *Module) Fulfill() {
	//模块其他详细信息
	m.KVs = make(map[string]string)
	for _, kv := range m.Data.M {
		key, errK := kv["key"]
		val, errV := kv["value"]
		if !errK || !errV || key == "" || val == "" {
			tlog.Handler.Errorf(context.Background(), tlog.DLTagUndefined, "Failed to read module info in key-value style!")
			continue
		}
		if key == "language" && val != "go" {
			delete(Module2Host, m.Name)
			delete(Host2Module, val)
			return
		}
		if key == "context" {
			Module2Host[m.Name] = val
			Host2Module[val] = m.Name
		}
		if m.KVs == nil {
			m.KVs = make(map[string]string)
		}
		m.KVs[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}
	//读取自配置时，为了减少配置复杂度，代码里会设置一些默认信息
	m.KVs["language"] = "go"
	m.KVs["protocol"] = "http"
	//模块名
	if _, ok := ModuleNamesUniq[m.Name]; !ok {
		ModuleNamesUniq[m.Name] = 1
		ModuleNames = append(ModuleNames, Name{m.Name})
	}
}

func (r *Maps) UnmarshalJSON(data []byte) error {
	quoted, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(quoted), &r.M)
}
