package response

// Pagination metadata for responses
type Pagination struct {
	CurrentPage  int  `json:"currentPage"`
	NextPage     *int `json:"nextPage"`
	PrevPage     *int `json:"prevPage"`
	TotalPage    int  `json:"totalPage"`
	TotalRecords int  `json:"totalRecords"`
}

// GlobalResponse is the standard format required by rules
type GlobalResponse struct {
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	ReqID      string      `json:"reqId"`
	Status     string      `json:"status"` // "T" or "F"
}

// ErrorResponse is for 4xx and 5xx errors
type ErrorResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
