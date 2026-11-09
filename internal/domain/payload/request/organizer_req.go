package request

type UpsertOrganizerProfileRequest struct {
	OrgName     string `json:"orgName"`
	Description string `json:"description"`
	Phone       string `json:"phone"`
	SocialMedia string `json:"socialMedia"`
}

type GetOrganizerStaffListRequest struct {
	Page     int
	PageSize int
	Search   string
	Status   string
}
