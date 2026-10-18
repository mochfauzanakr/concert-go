package usecase

import (
	"context"

	"concert-go/internal/domain/payload/response"
	"concert-go/internal/repository"
)

type RoleUsecase struct {
	roleRepo repository.RoleRepository
}

func NewRoleUsecase(roleRepo repository.RoleRepository) *RoleUsecase {
	return &RoleUsecase{roleRepo: roleRepo}
}

func (u *RoleUsecase) GetRoles(ctx context.Context) ([]response.RoleResponse, error) {
	roles, err := u.roleRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var res []response.RoleResponse
	for _, r := range roles {
		res = append(res, response.RoleResponse{
			ID:   r.ID,
			Name: r.Name,
		})
	}
	return res, nil
}
