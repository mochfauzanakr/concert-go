package handler

import (
	"konserGo/internal/service"
	"konserGo/internal/pkg/util"

	"github.com/gin-gonic/gin"
)

type HelloHandler interface {
	SayHello(c *gin.Context)
}

type helloHandlerImpl struct {
	helloService service.HelloService
}

// NewHelloHandler implements HelloHandler
func NewHelloHandler(helloService service.HelloService) HelloHandler {
	return &helloHandlerImpl{
		helloService: helloService,
	}
}

func (h *helloHandlerImpl) SayHello(c *gin.Context) {
	msg, err := h.helloService.GetGreeting()
	if err != nil {
		util.RespondInternalError(c)
		return
	}

	util.RespondOK(c, "Success greeting", map[string]string{"greeting": msg}, nil)
}
