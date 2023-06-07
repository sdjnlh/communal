package app

import (
	"code.letsit.cn/go/common"
	"encoding/json"
	"fmt"

	"code.letsit.cn/go/common/log"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	StarterLog        = "LOG"
	ConfigFileNameLog = "log_config"
)

type LogStarter struct {
	*BaseStarter
}

func (starter *LogStarter) Start(ctx *common.Context) error {
	var logConfig zap.Config
	var conf *viper.Viper = viper.New()
	var err error

	err = LoadConfig(ConfigFileNameLog, conf)
	if err != nil {
		return err
	}
	m := map[string]interface{}{}
	if err = conf.UnmarshalKey("log", &m); err != nil {
		return err
	}

	logcfgs, _ := json.Marshal(m)
	fmt.Println("log config: \n" + string(logcfgs))

	if err := json.Unmarshal(logcfgs, &logConfig); err != nil {
		return err
	}

	if log.Logger.Logger, err = logConfig.Build(); err != nil {
		return err
	}
	log.Slog = log.Logger.Sugar()

	log.Logger.Info("logger inited")

	return nil
}

func init() {
	RegisterStarter(&LogStarter{BaseStarter: NewBaseStarter(StarterLog, PriorityHigh+100)})
}
