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
