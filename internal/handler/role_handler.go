package handler

import (
	"concert-go/internal/usecase"
	"concert-go/internal/util"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	roleUsecase *usecase.RoleUsecase
}

func NewRoleHandler(roleUsecase *usecase.RoleUsecase) *RoleHandler {
	return &RoleHandler{roleUsecase: roleUsecase}
}

func (h *RoleHandler) GetRoles(c *gin.Context) {
	roles, err := h.roleUsecase.GetRoles(c.Request.Context())
	if err != nil {
		util.RespondError(c, err)
		return
	}

	util.RespondOK(c, "Roles fetched successfully", roles, nil)
}
