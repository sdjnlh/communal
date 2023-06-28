package app

import (
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/sdjnlh/communal"
	"github.com/sdjnlh/communal/config"
	"github.com/spf13/viper"
	"xorm.io/xorm"
)

type App interface {
	Starter
	Mount(app ...App) App
	SetMaster(master App)
	Master() App
	IsMaster() bool
}

type BaseApp struct {
	Configurator
	name     string
	master   App
	Mounts   *[]App
	isMaster bool

	Rpc     bool
	modules []communal.IModule
	DB      *xorm.Engine
	Redis   *redis.Client
}

func (app *BaseApp) SetMaster(master App) {
	app.master = master
}

func (app *BaseApp) Master() App {
	return app.master
}

func (app *BaseApp) IsMaster() bool {
	return app.isMaster
}

func (app *BaseApp) SetConfigFileName(name string) App {
	app.Configurator.FileName = name
	return app
}

func (app *BaseApp) Mount(apps ...App) App {
	app.isMaster = true
	if app.Mounts == nil {
		app.Mounts = &[]App{}
	}
	for _, ap := range apps {
		*app.Mounts = append(*app.Mounts, ap)
		if ap.Priority() > app.priority {
			ap.SetPriority(app.priority - 1)
		}
		ap.SetMaster(app)
	}

	return app
}

func (app *BaseApp) Name() string {
	return app.name
}

func (app *BaseApp) Register(modules ...communal.IModule) {
	app.modules = append(app.modules, modules...)
}

func (app *BaseApp) SetRedisConnection(client *redis.Client) {
	app.Redis = client
}

func (app *BaseApp) Start(ctx *communal.Context) error {
	//app.Configurator.BaseStarter = *(NewBaseStarter(app.name+"_config", PriorityHigh))
	(&app.Configurator).SetApp(app)
	//RegisterStarter(&app.Configurator)
	err := (&app.Configurator).Start(ctx)
	if err != nil {
		return err
	}

	dbn := app.RawConfig.GetString(app.name + ".db")

	for _, module := range app.modules {
		if module.DbEnabled() {
			if module.GetDbName() == "" {
				module.SetDbName(dbn)
			}

			if dbn == "" {
				panic("db enabled for module " + module.GetName() + ", but no name specified for module or app " + app.name)
			}

			if ctx.Get("db."+dbn) == nil {
				ListenDB(module)
			} else {
				module.SetDB(ctx.Get("db." + dbn).(*xorm.Engine))
			}
		}
	}
	RegisterStarter(&DbStarter{
		BaseStarter: BaseStarter{
			name:     app.Name() + ".DB",
			priority: PriorityMiddle,
		},
		Namespace: app.name,
		//DbHolder:  app,
	})

	fmt.Println(app.name + ": service db starter registered")
	RegisterStarter(&RedisStarter{
		BaseStarter: BaseStarter{
			name:     app.Name() + ".REDIS",
			priority: PriorityMiddle,
		},
		Namespace:   app.name,
		RedisHolder: app,
	})
	fmt.Println(app.name + ": service redis starter registered")

	if app.isMaster && app.Mounts != nil {
		fmt.Println("register mounts")
		for _, mnt := range *app.Mounts {
			RegisterStarter(mnt)
		}
	}

	return nil
}

type Configurator struct {
	BaseStarter
	FileName     string
	RawConfig    *viper.Viper
	Subscription []config.Pair
}

func (configurator *Configurator) Subscribe(key string, target interface{}) {
	configurator.Subscription = append(configurator.Subscription, config.Pair{Key: key, Target: target})
}

/**
* 1, try to load config with config file name if given
* 2, try to load config with app name
* 3, master APP:
*	a, error if not exist
*	b,
 */
func (configurator *Configurator) Start(ctx *communal.Context) error {
	fileName := configurator.FileName
	if fileName == "" {
		fileName = configurator.app.Name()
	}
	configurator.RawConfig = viper.New()
	err := LoadConfig(fileName, configurator.RawConfig)
	if err != nil {
		//fmt.Printf("fail to read config file \t%s\t%v \n", fileName, err)
		if configurator.app.IsMaster() {
			return errors.New("fail to read master config file")
		} else {
			configurator.RawConfig = ctx.Get(configurator.app.Master().Name() + ".config").(*viper.Viper)
		}
	}

	//if configurator.app.IsMaster() {
	//	config.Config = configurator.RawConfig
	//}

	ctx.Set(configurator.app.Name()+".config", configurator.RawConfig)

	if configurator.RawConfig != nil && configurator.Subscription != nil {
		for _, pair := range configurator.Subscription {
			//if configurator.RawConfig == nil {
			//	fmt.Println("unmarshal config: ", pair.Key, nil)
			//} else {
			fmt.Println("unmarshal config: ", pair.Key, configurator.RawConfig.Get(pair.Key))
			//}

			err = configurator.RawConfig.UnmarshalKey(pair.Key, pair.Target)
			if err != nil {
				return err
			}
			//fmt.Println(pair.Target)
		}
	}

	return nil
}
