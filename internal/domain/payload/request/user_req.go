package request

type InviteUserRequest struct {
	Email string `json:"email" binding:"required,email,max=100"`
	Role  string `json:"role" binding:"required"`
}

type AcceptInviteRequest struct {
	Token           string `json:"token"`
	Name            string `json:"name" binding:"required,min=2,max=100"`
	Password        string `json:"password" binding:"required,min=8,max=128"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,min=8,max=128"`
}

type GetUserListRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Search   string `json:"search"`
	RoleID   *int   `json:"roleId"`
	Status   string `json:"status"`
}

type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type UpdateProfileRequest struct {
	Name string `form:"name" json:"name" binding:"omitempty,min=2,max=100"`
}

type AdminUpdateUserRequest struct {
	Name   string `json:"name" binding:"required,min=2,max=100"`
	RoleID int    `json:"roleId" binding:"required"`
}
