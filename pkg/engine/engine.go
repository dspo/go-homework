package engine

import (
	"github.com/gin-gonic/gin"
)

type Engine struct {
	*gin.Engine
}

func New() *Engine {
	return &Engine{
		Engine: gin.New(),
	}
}
