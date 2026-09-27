package handler

import (
	"concert-go/internal/usecase"
	"concert-go/internal/util"

	
	"github.com/gin-gonic/gin"
)

type HelloHandler interface {
	SayHello(c *gin.Context)
}

type helloHandlerImpl struct {
	helloUsecase usecase.HelloUsecase
}

// NewHelloHandler implements HelloHandler
func NewHelloHandler(helloUsecase usecase.HelloUsecase) HelloHandler {
	return &helloHandlerImpl{
		helloUsecase: helloUsecase,
	}
}

func (h *helloHandlerImpl) SayHello(c *gin.Context) {
	msg, err := h.helloUsecase.GetGreeting()
	if err != nil {
		util.RespondInternalError(c)
		return
	}

	util.RespondOK(c, "Success greeting", map[string]string{"greeting": msg}, nil)
}

