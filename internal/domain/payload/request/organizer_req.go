package request

import "concert-go/internal/domain/model"

type UpsertOrganizerProfileRequest struct {
	OrgName     string `json:"orgName"`
	Description string `json:"description"`
	Phone       string `json:"phone"`
	SocialMedia string `json:"socialMedia"`
}

type GetOrganizerStaffListRequest struct {
	model.BaseListParams
	Status string `json:"status"`
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
	model.BaseListParams
	Status string `json:"status"`
}

type ReviewOrganizerApplicationRequest struct {
	Status          string `json:"status"`
	RejectionReason string `json:"rejectionReason"`
}
