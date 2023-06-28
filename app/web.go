package app

import (
	"github.com/gin-gonic/gin"
	"github.com/sdjnlh/communal"
	"github.com/sdjnlh/communal/log"
	"go.uber.org/zap"
	"html/template"
	"net/http"
)

type Web struct {
	BaseApp

	Domain           string
	LoginUrl         string
	StateManager     *StateManager
	Port             string
	HtmlTemplateRoot string

	HtmlTemplate        *template.Template
	HtmlTemplateFuncMap template.FuncMap
}

func NewWeb(name string) *Web {
	web := &Web{
		BaseApp: BaseApp{
			name: name,
		},
	}
	web.SetPriority(PriorityHigh)
	return web
}

func (app *Web) Start(cxt *communal.Context) error {
	app.Subscribe(app.name, app)
	err := (&app.BaseApp).Start(cxt)
	if err != nil {
		return err
	}

	if app.HtmlTemplateRoot != "" {
		//RegisterStarter(HtmlTemplateStarter{
		//	BaseStarter: NewBaseStarter(app.name + "_html_template"),
		//	RootDir: app.HtmlTemplateRoot,
		//})
		//app.HtmlTemplate = template.Must(template.New("").Funcs(app.HtmlTemplateFuncMap).ParseGlob(app.HtmlTemplateRoot + "/*.html"))
	}

	log.Logger.Debug("add ustm: "+app.name+".USTM", zap.Any("Name()", app.Name()))
	RegisterStarter(&StateManagerStarter{
		BaseStarter: &BaseStarter{
			name:     app.name + ".USTM",
			priority: PriorityLow,
		},
		Namespace:          app.name,
		StateManagerHolder: &app.StateManager,
	})

	return nil
}

func (web *Web) HtmlUserInterceptor(c *gin.Context) {
	log.Logger.Debug("HTML User Interceptor: check user id in context", zap.Int64("id", c.GetInt64(communal.UserIdKey)))
	if c.GetInt64(communal.UserIdKey) <= 0 {
		c.Redirect(http.StatusTemporaryRedirect, web.LoginUrl)
		return
	}

	log.Logger.Debug("user has login", zap.Int64("id", c.GetInt64(communal.UserIdKey)))
	c.Next()
}

func (app *Web) ManageUserState(engine *gin.Engine, builder StateBuilder) {
	if app.StateManager == nil {
		panic("no state manager defined for web " + app.Name())
	}

	app.StateManager.Store.Use(engine)
	app.StateManager.Store.SetStateBuilder(builder)
}
