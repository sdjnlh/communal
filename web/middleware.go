package web

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sdjnlh/communal"
	"github.com/sdjnlh/communal/app"
	"github.com/sdjnlh/communal/log"
	"go.uber.org/zap"
)

// func CorsHandler(conf CorsConfig) func(c *gin.Context) {
func CorsHandler(webapp *app.Web) func(c *gin.Context) {
	// cc = conf
	if webapp.Started() {
		if webapp.RawConfig != nil {
			_ = webapp.RawConfig.UnmarshalKey("cors", &cc)
		}
	} else {
		webapp.Subscribe("cors", &cc)
	}

	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("origin")
		log.Logger.Debug(c.Request.RequestURI + "   origin: " + origin + "   referer: " + c.Request.Referer())

		if !cc.AllowAll && !strings.HasSuffix(origin, cc.AllowOrigin) {
			//todo use referer for request source checking maybe, but not origin
			//todo log this request, it's suspicious
			log.Logger.Warn("invalid request from "+origin, zap.String("allow origin", cc.AllowOrigin))
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		//Logger.Debug("request from ", zap.String("origin", origin))

		if strings.ToUpper(method) == "OPTIONS" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept,Authorization,Cache-Control,Content-Type,DNT,If-Modified-Since,Keep-Alive,Origin,User-Agent,X-Mx-ReqToken,X-Requested-With")
			c.Writer.Header().Set("Access-Control-Max-Age", "1728000")
			c.AbortWithStatus(204)
			return
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Next()
		return
	}
}

type CorsConfig struct {
	AllowAll    bool
	AllowOrigin string
}

var cc CorsConfig

var UserInterceptor = func(c *gin.Context) {
	v, ok := c.Get(communal.UserKey)
	if ok {
		if v.(communal.IdInf).GetId() > 0 {
			c.Next()
			return
		}
	}

	c.AbortWithStatus(http.StatusUnauthorized)
}
