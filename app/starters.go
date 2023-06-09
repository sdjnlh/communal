package app

import (
	"errors"
	"html/template"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/sdjnlh/communal"
	"github.com/sdjnlh/communal/log"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"xorm.io/xorm"
)

type DbHolder interface {
	SetDb(*xorm.Engine)
}

type DbStarter struct {
	BaseStarter
	Namespace string
	DbHolder  DbHolder
	listeners map[string][]communal.DBListener
}

var dbListeners map[string][]communal.DBListener

func ListenDB(listeners ...communal.DBListener) {
	if dbListeners == nil {
		dbListeners = map[string][]communal.DBListener{}
	}

	for _, listener := range listeners {
		if listener.DbEnabled() {
			dbListeners[listener.GetDbName()] = append(dbListeners[listener.GetDbName()], listener)
		}
	}
}

func (starter *DbStarter) Start(ctx *communal.Context) error {
	cfg := ctx.MustGet(starter.Namespace + ".config").(*viper.Viper)

	dbns := cfg.GetStringMap("db")
	if len(dbns) == 0 {
		log.Slog.Warn("no db config found for db starter ", starter.name)
		return nil
	}

	for dbn, _ := range dbns {
		var conn *xorm.Engine
		var err error

		if ctx.Get("db."+dbn) == nil {
			conn, err = BuildDBConnection(cfg.Sub("db." + dbn))
			if err != nil {
				return err
			}
			ctx.Set("db."+dbn, conn)

			if len(dbListeners) == 0 || len(dbListeners[dbn]) == 0 {
				continue
			}
			for _, listener := range dbListeners[dbn] {
				listener.SetDB(conn)
			}
		}
	}

	return nil
}

type RedisHolder interface {
	SetRedisConnection(*redis.Pool)
}

type RedisStarter struct {
	BaseStarter
	Namespace   string
	RedisHolder RedisHolder
}

func (starter *RedisStarter) Start(ctx *communal.Context) error {
	cfg := ctx.MustGet(starter.Namespace + ".config").(*viper.Viper)

	dbn := cfg.GetString(starter.Namespace + ".redis")
	if dbn == "" {
		return nil
	}

	var conn *redis.Pool
	var err error
	if ctx.Get("redis."+dbn) == nil {
		conn, err = BuildRedisConnection(cfg.Sub("redis." + dbn))
		if err != nil {
			return err
		}
		ctx.Set("redis."+dbn, conn)
	} else {
		conn = ctx.Get("redis." + dbn).(*redis.Pool)
	}

	starter.RedisHolder.SetRedisConnection(conn)

	return nil
}

type dbConfig struct {
	Clustered bool
	Name      string
	Ref       string
	Type      string
	Uri       string
	MaxIdle   int
	MaxOpen   int
	ShowSQL   bool
}

func BuildDBConnection(config *viper.Viper) (*xorm.Engine, error) {
	conf := dbConfig{}
	err := config.Unmarshal(&conf)
	if err != nil {
		return nil, err
	}
	if conf.Clustered {
		//TODO build by cluster config

		return nil, nil
	}

	engine, err := xorm.NewEngine(conf.Type, conf.Uri)
	if err != nil {
		return engine, err
	}

	if conf.MaxIdle > 0 {
		engine.SetMaxIdleConns(conf.MaxIdle)
	}

	if conf.MaxOpen > 0 {
		engine.SetMaxOpenConns(conf.MaxOpen)
	}

	engine.ShowSQL(conf.ShowSQL)

	return engine, err
}

type redisConfig struct {
	MaxIdle     int
	IdleTimeout int
	Server      string
	Auth        bool
	Password    string
}

func BuildRedisConnection(config *viper.Viper) (*redis.Pool, error) {
	if config == nil {
		return nil, errors.New("Nil config when build redis connection")
	}
	conf := redisConfig{}
	err := config.Unmarshal(&conf)
	if err != nil {
		return nil, err
	}

	log.Logger.Debug("redis", zap.Any("config", conf))
	return &redis.Pool{
		MaxIdle:     conf.MaxIdle,
		IdleTimeout: time.Second * time.Duration(conf.IdleTimeout),
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", conf.Server)
			if err != nil {
				return nil, err
			}

			if conf.Auth {
				if _, err := c.Do("AUTH", conf.Password); err != nil {
					_ = c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}, nil
}

type HtmlTemplateStarter struct {
	*BaseStarter
	RootDir             string
	HtmlTemplateHolder  **template.Template
	HtmlTemplateFuncMap template.FuncMap
}

func (starter *HtmlTemplateStarter) Start() (err error) {
	if starter.RootDir == "" {
		return errors.New("no template root")
	}

	*starter.HtmlTemplateHolder = template.Must(template.New("").Funcs(starter.HtmlTemplateFuncMap).ParseGlob(starter.RootDir + "/*.html"))
	return nil
}
