package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthzHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	}
}

func NotYetImplemented(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotImplemented)
}
