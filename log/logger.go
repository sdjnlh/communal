package log

import (
	"encoding/json"
	"fmt"
	"github.com/sdjnlh/communal/config"
	"go.uber.org/zap"
	"log"
	"sync"
)

type logger struct {
	*zap.Logger
}

func (l *logger) ErrorE(err error, msg ...interface{}) {
	var s string
	if len(msg) > 0 {
		if fmt.Sprintf("%T", msg[0]) == "string" {
			s = msg[0].(string)
		}
	}
	l.Error(s, zap.Error(err))
}

var (
	Logger = logger{}
	Slog   *zap.SugaredLogger
)

var (
	l    *logger
	once sync.Once
)

func newlogger() *logger {
	once.Do(func() {

		cg := config.Configer()

		m := map[string]interface{}{}
		if err := cg.UnmarshalKey("log", &m); err != nil {
			log.Fatal(err)
			return
		}

		var logConfig zap.Config
		cgmap, _ := json.Marshal(m)
		if err := json.Unmarshal(cgmap, &logConfig); err != nil {
			log.Fatal(err)
			return
		}
		l = &logger{}
		var err error
		l.Logger, err = logConfig.Build()
		if err != nil {
			log.Fatal(err)
			return
		}

		l.Info("log init success")
	})
	return l
}
