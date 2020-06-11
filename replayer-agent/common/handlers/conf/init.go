package conf

import (
	"fmt"
	//"path/filepath"
	"os"
	//"runtime"

	"github.com/spf13/viper"
)

var Handler *viper.Viper
var HandlerInfo *viper.Viper
var newConfs map[string]*viper.Viper

// Root current dir
var Root string

func Init(confPath string) {
	var err error
	newConfs = make(map[string]*viper.Viper)
	//_, fn, _, _ := runtime.Caller(0)
	//Root = filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(fn))))
	Root, err = os.Getwd()
	if err != nil {
		panic(fmt.Errorf("Initialize Root error: %s", err))
	}

	Handler = LoadConfig(confPath)
	HandlerInfo = LoadConfig(Root + "/conf/info.toml")
}

func LoadConfig(confPath string) *viper.Viper {
	handler := viper.New()

	if confPath == "" {
		handler.SetConfigName("app") // 默认文件配置文件为app.toml
		handler.AddConfigPath(Root + "/conf")
	} else {
		handler.SetConfigFile(confPath)
	}
	err := handler.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	return handler
}

// FreshHandler 从磁盘读取配置文件，方便本地修改配置的测试，无需重启服务
func FreshHandler() {
	Handler.ReadInConfig()
	HandlerInfo.ReadInConfig()
}

// NewConf 实例化读取配置文件
func NewConf(confFile string) *viper.Viper {
	if obj, ok := newConfs[confFile]; ok {
		obj.ReadInConfig()
		return obj
	} else {
		viperObj := viper.New()
		viperObj.SetConfigFile(confFile)
		newConfs[confFile] = viperObj
		viperObj.ReadInConfig()
		return viperObj
	}

}
