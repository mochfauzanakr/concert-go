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

type ApplyOrganizerRequest struct {
	PicName     string `json:"picName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	OrgName     string `json:"orgName"`
	SocialMedia string `json:"socialMedia"`
	Description string `json:"description"`
}

type GetOrganizerApplicationListRequest struct {
	Page     int
	PageSize int
	Search   string
	Status   string
}

type ReviewOrganizerApplicationRequest struct {
	Status          string `json:"status"`
	RejectionReason string `json:"rejectionReason"`
}
