package router

import "github.com/gin-gonic/gin"

type irouter interface {
	Register(*gin.RouterGroup)
}

func Register(root *gin.RouterGroup, child irouter) {
	child.Register(root)
}
