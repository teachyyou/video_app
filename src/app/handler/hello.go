package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type HelloHandler struct{}

func NewHelloHandler() *HelloHandler {
	return &HelloHandler{}
}

func (h *HelloHandler) Hello(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Hello World!"})
}

var HelloModule = fx.Module("hello-handler", fx.Provide(NewHelloHandler))
