package application

import (
	"github.com/gin-gonic/gin"

	"github.com/dspo/go-homework/internal/handler"
	"github.com/dspo/go-homework/pkg/engine"
)

type Context struct {
	Engine *engine.Engine

	GET, POST, PUT, PATCH, DELETE, Any func(string, ...gin.HandlerFunc) gin.IRoutes
}

func NewContext(gn *engine.Engine) Context {
	return Context{
		Engine: gn,
		GET:    wrap(gn.GET),
		POST:   wrap(gn.POST),
		PUT:    wrap(gn.PUT),
		PATCH:  wrap(gn.PATCH),
		DELETE: wrap(gn.DELETE),
		Any:    wrap(gn.Any),
	}
}

func wrap(f func(string, ...gin.HandlerFunc) gin.IRoutes) func(string, ...gin.HandlerFunc) gin.IRoutes {
	return func(s string, handlerFunc ...gin.HandlerFunc) gin.IRoutes {
		if len(handlerFunc) == 0 {
			handlerFunc = append(handlerFunc, handler.NotYetImplemented)
		}
		return f(s, handlerFunc...)
	}
}
