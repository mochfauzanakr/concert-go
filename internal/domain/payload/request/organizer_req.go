package request

type UpsertOrganizerProfileRequest struct {
	OrgName     string `json:"orgName"`
	Description string `json:"description"`
	Phone       string `json:"phone"`
	SocialMedia string `json:"socialMedia"`
}
