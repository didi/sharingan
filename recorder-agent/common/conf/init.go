package conf

import (
	"fmt"

	"github.com/didi/sharingan/recorder-agent/common/path"

	"github.com/spf13/viper"
)

// Handler Handler
var Handler *viper.Viper

func init() {
	confPath := path.Root + "/conf/app.toml"
	Handler = LoadConfig(confPath)
}

// LoadConfig LoadConfig
func LoadConfig(confPath string) *viper.Viper {
	handler := viper.New()
	handler.SetConfigFile(confPath)

	if err := handler.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	return handler
}
