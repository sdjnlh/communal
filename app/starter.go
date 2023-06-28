package app

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"

	"github.com/sdjnlh/communal"
)

const (
	PriorityHighest = 1000
	PriorityHigh    = 900
	PriorityMiddle  = 600
	PriorityLow     = 300
	PriorityLowest  = 0
)

type Starter interface {
	Name() string
	Priority() int
	SetPriority(int) Starter
	SetAppName(appName string) Starter
	SetApp(app App) Starter
	AppName() string
	Start(ctx *communal.Context) error
	Started() bool
	SetStarted(bool) Starter
}

type BaseStarter struct {
	name     string
	priority int
	started  bool
	appName  string
	app      App
	action   func(ctx *communal.Context) error
}

func NewBaseStarter(name string, priority int) *BaseStarter {
	return &BaseStarter{
		name:     name,
		priority: priority,
	}
}

func (base *BaseStarter) Name() string {
	return base.name
}

func (base *BaseStarter) Priority() int {
	return base.priority
}

func (base *BaseStarter) SetPriority(priority int) Starter {
	base.priority = priority
	return base
}

func (base *BaseStarter) SetAppName(appName string) Starter {
	base.appName = appName
	return base
}

func (base *BaseStarter) AppName() string {
	return base.appName
}

func (base *BaseStarter) Started() bool {
	return base.started
}

func (base *BaseStarter) SetStarted(started bool) Starter {
	base.started = started
	return base
}

func (base *BaseStarter) SetApp(app App) Starter {
	base.app = app
	return base
}

func (base *BaseStarter) Action(action func(ctx *communal.Context) error) Starter {
	base.action = action
	return base
}

func (base *BaseStarter) Start(ctx *communal.Context) error {
	if base.action != nil {
		return base.action(ctx)
	}

	return nil
}

var (
	controller = &StartController{}
)

func RegisterStarter(starter Starter) {
	fmt.Println("Register starter >> " + starter.Name() + " [" + reflect.TypeOf(starter).String() + "]")
	controller.register(starter)
}

func Start() error {
	controller.ctx = communal.Context{}
	err := controller.startNext()
	controller = nil
	return err
}

type StartListener func(ctx communal.Context) error

func OnStarted(starterName string, listener StartListener) {
	if controller.listenersMap == nil {
		controller.listenersMap = map[string][]StartListener{}
	}
	controller.listenersMap[starterName] = append(controller.listenersMap[starterName], listener)
}

func init() {}

type StartController struct {
	ctx           communal.Context
	mu            sync.RWMutex
	startersMap   map[string]Starter
	startersArray []Starter
	listenersMap  map[string][]StartListener
}

func (controller *StartController) register(starter Starter) {
	//fmt.Println("Register starter: " + starter.Name())
	controller.mu.Lock()

	if controller.startersMap == nil {
		controller.startersMap = make(map[string]Starter)
	}

	//printStarters(starter.Name(), controller.startersArray)
	if controller.startersMap[starter.Name()] == nil {
		controller.startersMap[starter.Name()] = starter
		var arr []Starter
		var added = false
		for _, str := range controller.startersArray {
			if starter.Priority() > str.Priority() && !added {
				arr = append(arr, starter)
				added = true
			}
			arr = append(arr, str)
		}

		if !added {
			arr = append(arr, starter)
		}
		controller.startersArray = arr
	} else {
		panic("duplicated starter startersArray: " + starter.Name())
	}
	//printStarters(starter.Name(), controller.startersArray)
	controller.mu.Unlock()
}

func printStarters(prefix string, starts []Starter) {
	str := ""
	for index, st := range controller.startersArray {
		if index > 0 {
			str += ","
		}
		str += st.Name() + ":" + strconv.Itoa(st.Priority())
	}
	fmt.Println(prefix + "-----" + str)
}

func (controller *StartController) startNext() error {
	var err error
	var starter Starter
	controller.mu.Lock()
	if len(controller.startersArray) == 0 {
		return nil
	}

	starter = controller.startersArray[0]
	if len(controller.startersArray) > 1 {
		controller.startersArray = controller.startersArray[1:]
		printStarters(starter.Name(), controller.startersArray)
	} else {
		controller.startersArray = []Starter{}
	}

	controller.mu.Unlock()

	err = controller.startStarter(starter)
	if err != nil {
		fmt.Println("fail to start starter " + starter.Name())
		fmt.Println(err.Error())
		return err
	}

	return controller.startNext()
}

func (controller *StartController) startStarter(starter Starter) error {
	if starter.Started() {
		panic("starter " + starter.Name() + " has been started")
	}
	err := starter.Start(&controller.ctx)
	if err != nil {
		return err
	} else {
		starter.SetStarted(true)

		listeners := controller.listenersMap[starter.Name()]

		if listeners != nil {
			for _, listener := range listeners {
				if err = listener(controller.ctx); err != nil {
					return err
				}
			}
		}
		fmt.Println("Starter started << " + starter.Name())
		return nil
	}
}
