package app

import (
	"flag"
	"fmt"

	"github.com/sdjnlh/communal"

	"github.com/spf13/viper"
)

var confRoot *string

/**
*	load config from file
 */
type ConfigLoader struct {
	*BaseStarter
	ConfigFileName string
	Config         *viper.Viper
}

func (cs *ConfigLoader) Start(ctx communal.Context) error {
	return LoadConfig(cs.ConfigFileName, cs.Config)
}

func LoadConfig(name string, config *viper.Viper) error {
	fmt.Println("load config file " + name)
	flag.Parse()
	config.SetConfigName(name)
	config.AddConfigPath(*confRoot)
	config.AddConfigPath("$HOME/.letsit.cn/")
	config.AddConfigPath(".")
	config.AddConfigPath("./conf")
	err := config.ReadInConfig()
	if err != nil {
		fmt.Printf("Fatal error config file: %s \n", err)
	}

	return err
}

func init() {
	confRoot = flag.String("conf-dir", "/etc/letsit.cn/", "config root dir")
}
