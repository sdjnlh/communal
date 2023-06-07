package app

import (
	"encoding/json"
	be "errors"
	"net/http"

	"code.letsit.cn/go/common"
	"code.letsit.cn/go/common/log"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	redisDriver "github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	StateStoreTypeJwt     = "jwt"
	StateStoreTypeSession = "session"
)

type StateManager struct {
	Store StateStore
}

func (manager *StateManager) SetAll(c *gin.Context, values map[string]interface{}) error {
	if manager.Store == nil {
		return be.New("no state store provided")
	}

	return manager.Store.SetAll(c, values)
}

func (manager *StateManager) ClearAll(c *gin.Context, values map[string]interface{}) error {
	if manager.Store == nil {
		return be.New("no state store provided")
	}

	return manager.Store.ClearAll(c, values)
}

func (manager *StateManager) SetUser(c *gin.Context, user map[string]interface{}) error {
	if manager.Store == nil {
		return be.New("no state store provided")
	}

	return manager.Store.SetUser(c, user)
}

func (manager *StateManager) Domain() string {
	if manager.Store == nil {
		return ""
	}
	return manager.Store.Domain()
}

func (manager *StateManager) Path() string {
	if manager.Store == nil {
		return ""
	}
	return manager.Store.Path()
}

func (manager *StateManager) MaxAge() int {
	if manager.Store == nil {
		return 0
	}
	return manager.Store.MaxAge()
}

type StateManagerStarter struct {
	*BaseStarter
	Namespace          string
	StateManagerHolder **StateManager
}

func (starter *StateManagerStarter) Start(ctx *common.Context) error {
	log.Logger.Debug("start state manager")
	cfg := ctx.MustGet(starter.Namespace + ".config").(*viper.Viper)

	cfgValue := cfg.GetString(starter.Namespace + ".ustm")
	if cfgValue == "" {
		return nil
	}

	var key = "ustm." + cfgValue
	if ctx.Get(key) == nil {
		var err error
		*starter.StateManagerHolder, err = NewStateManager(cfg.Sub(key), ctx)
		if err != nil {
			return err
		}
		ctx.Set(key, *starter.StateManagerHolder)
	} else {
		*starter.StateManagerHolder = ctx.Get(key).(*StateManager)
	}

	return nil
}

func NewStateManager(config *viper.Viper, ctx *common.Context) (manager *StateManager, err error) {
	tp := config.GetString("type")

	var store StateStore
	if tp == StateStoreTypeJwt {
		options := JwtOptions{}
		err := config.Unmarshal(&options)
		if err != nil {
			log.Logger.Warn("fail to build jwt state store", zap.Error(err))
			return nil, err
		}
		store = &JwtStore{Options: options}
	} else if tp == StateStoreTypeSession {
		options := SessionOptions{}
		err := config.Unmarshal(&options)
		if err != nil {
			log.Logger.Warn("fail to build session state store", zap.Error(err))
			return nil, err
		}
		err = config.Unmarshal(&options.Options)
		if err != nil {
			log.Logger.Warn("fail to build session state store", zap.Error(err))
			return nil, err
		}
		store, err = newSessionStore(options, ctx)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, be.New("invalid state manager type " + tp)
	}

	return &StateManager{Store: store}, nil
}

func (manager *StateManager) ParseUser() func(c *gin.Context) {
	if manager.Store == nil {
		return nil
	}
	return manager.Store.ParseUser
}

type StateBuilder func(map[string]interface{}, *gin.Context) interface{}
type StateStore interface {
	Use(engine *gin.Engine)
	ParseUser(c *gin.Context)
	SetStateBuilder(StateBuilder)
	SetUser(*gin.Context, map[string]interface{}) error
	//Get(c *gin.Context, key string) interface{}
	//GetAll(c *gin.Context, keys ...string) map[string]interface{}
	Set(c *gin.Context, key string, value interface{}) error
	SetAll(*gin.Context, map[string]interface{}) error
	ClearAll(*gin.Context, map[string]interface{}) error
	Domain() string
	Path() string
	MaxAge() int
}

type StoreBase struct {
	StateBuilder StateBuilder
}

func (base *StoreBase) SetStateBuilder(builder StateBuilder) {
	base.StateBuilder = builder
}

type JwtStore struct {
	StoreBase
	Options JwtOptions
}

type JwtOptions struct {
	Name     string
	Domain   string
	Path     string
	MaxAge   int
	Secure   bool
	HttpOnly bool
	Secret   string
	SameSite int
}

func (js *JwtStore) Use(engine *gin.Engine) {
	engine.Use(js.ParseUser)
}

func (js *JwtStore) Get(key string) {

}

func (js *JwtStore) Set(c *gin.Context, key string, value interface{}) error {
	return nil
}

func (js *JwtStore) Domain() string {
	return js.Options.Domain
}

func (js *JwtStore) Path() string {
	return js.Options.Path
}

func (js *JwtStore) MaxAge() int {
	return js.Options.MaxAge
}

func (js *JwtStore) SetAll(c *gin.Context, values map[string]interface{}) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(values))

	// Sign and get the complete encoded token as a string using the secret
	//key := config.Config.GetString("jwt.key.public")
	tokenString, err := token.SignedString([]byte(js.Options.Secret))

	if err != nil {
		log.Logger.Error("token sign error", zap.String("error", err.Error()))
		return err
	}

	if js.Options.SameSite > 0 {
		c.SetSameSite(http.SameSiteNoneMode)
	}

	log.Logger.Debug("sign jwt token", zap.String("name", js.Options.Name), zap.String("token", tokenString))
	c.SetCookie(js.Options.Name, tokenString, js.Options.MaxAge, js.Options.Path, js.Options.Domain, js.Options.Secure, js.Options.HttpOnly)

	for key, value := range values {
		if value != nil {
			c.SetCookie(key, value.(string), js.Options.MaxAge, js.Options.Path, js.Options.Domain, js.Options.Secure, false)
		}
	}

	return nil
}

func (js *JwtStore) ClearAll(c *gin.Context, values map[string]interface{}) error {
	if js.Options.SameSite > 0 {
		c.SetSameSite(http.SameSiteNoneMode)
	}
	c.SetCookie(js.Options.Name, "", -1, js.Options.Path, js.Options.Domain, js.Options.Secure, js.Options.HttpOnly)

	for key, value := range values {
		if value != nil {
			c.SetCookie(key, "", -1, js.Options.Path, js.Options.Domain, js.Options.Secure, false)
		}
	}

	return nil
}

func (js *JwtStore) SetUser(c *gin.Context, user map[string]interface{}) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(user))

	// Sign and get the complete encoded token as a string using the secret
	//key := config.Config.GetString("jwt.key.public")
	tokenString, err := token.SignedString([]byte(js.Options.Secret))

	if err != nil {
		log.Logger.Error("token sign error", zap.String("error", err.Error()))
		return err
	}

	if js.Options.SameSite > 0 {
		c.SetSameSite(http.SameSiteNoneMode)
	}
	log.Logger.Debug("sign jwt token", zap.String("name", js.Options.Name), zap.String("token", tokenString))
	c.SetCookie(js.Options.Name, tokenString, js.Options.MaxAge, js.Options.Path, js.Options.Domain, js.Options.Secure, js.Options.HttpOnly)
	return nil
}

func (js *JwtStore) ParseUser(c *gin.Context) {
	tokenString, err := c.Cookie(js.Options.Name)
	log.Logger.Debug("jwt token", zap.String("name", js.Options.Name), zap.String("token", tokenString))
	if err != nil {
		log.Logger.Debug("failed to get token from cookie", zap.String("name", js.Options.Name), zap.Error(err))
		return
	}

	token, err := jwt.Parse(tokenString, js.key)
	if err != nil {
		log.Logger.Debug("failed to parse token", zap.String("token", tokenString), zap.Error(err))
		return
	}

	if js.StateBuilder != nil {
		js.StateBuilder(token.Claims.(jwt.MapClaims), c)
	}

	c.Next()
}

func (js *JwtStore) key(token *jwt.Token) (interface{}, error) {
	return []byte(js.Options.Secret), nil
}

type SessionStore struct {
	StoreBase
	Options   SessionOptions
	realStore sessions.Store
}

func newSessionStore(opt SessionOptions, ctx *common.Context) (store *SessionStore, err error) {
	store = &SessionStore{Options: opt}
	var realStore sessions.Store

	if opt.StoreType == "memory" {
		realStore = memstore.NewStore([]byte(opt.Secret))
	} else if opt.StoreType == "redis" {
		conn := ctx.Get("redis." + opt.Redis)
		if conn == nil {
			return nil, be.New("try to build session store with empty redis reference " + opt.Redis)
		}
		realStore, err = redis.NewStoreWithPool(conn.(*redisDriver.Pool), []byte(opt.Secret))
		if err != nil {
			return nil, err
		}
	}

	realStore.Options(opt.Options)
	store.realStore = realStore

	return
}

type SessionOptions struct {
	sessions.Options
	Type      string
	StoreType string
	Redis     string
	Name      string
	Secret    string
}

func (ss *SessionStore) Use(engine *gin.Engine) {
	engine.Use(sessions.Sessions(ss.Options.Name, ss.realStore), ss.ParseUser)
}

func (ss *SessionStore) ParseUser(c *gin.Context) {
	session := sessions.Default(c)
	userStr := session.Get(common.UserKey)
	if userStr != nil {
		userFields := map[string]interface{}{}
		err := json.Unmarshal(([]byte)(userStr.(string)), &userFields)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if ss.StateBuilder != nil {
			ss.StateBuilder(userFields, c)
		}
	}

	c.Next()
}

func (ss *SessionStore) SetUser(c *gin.Context, user map[string]interface{}) error {
	sess := sessions.Default(c)

	bts, err := json.Marshal(user)
	if err != nil {
		return err
	}
	sess.Set(common.UserKey, string(bts))

	return sess.Save()
}

func (ss *SessionStore) Set(c *gin.Context, key string, value interface{}) error {
	sess := sessions.Default(c)
	c.Set(key, value)
	return sess.Save()
}

func (ss *SessionStore) Domain() string {
	return ss.Options.Domain
}

func (ss *SessionStore) Path() string {
	return ss.Options.Path
}

func (ss *SessionStore) MaxAge() int {
	return ss.Options.MaxAge
}

func (ss *SessionStore) SetAll(c *gin.Context, values map[string]interface{}) error {
	sess := sessions.Default(c)

	for key, value := range values {
		sess.Set(key, value)
	}

	return sess.Save()
}

func (ss *SessionStore) ClearAll(c *gin.Context, values map[string]interface{}) error {
	sess := sessions.Default(c)

	for key, _ := range values {
		sess.Delete(key)
	}

	return sess.Save()
}
