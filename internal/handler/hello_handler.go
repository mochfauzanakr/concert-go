package handler

import (
	"concert-go/internal/usecase"
	"concert-go/internal/util"

	
	"github.com/gin-gonic/gin"
)

type HelloHandler struct {
	helloUsecase *usecase.HelloUsecase
}

// NewHelloHandler implements HelloHandler
func NewHelloHandler(helloUsecase *usecase.HelloUsecase) *HelloHandler {
	return &HelloHandler{
		helloUsecase: helloUsecase,
	}
}

func (h *HelloHandler) SayHello(c *gin.Context) {
	msg, err := h.helloUsecase.GetGreeting()
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Success greeting", map[string]string{"greeting": msg}, nil)
}

