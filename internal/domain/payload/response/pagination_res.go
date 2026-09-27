package response

// PaginationResponse provides pagination metadata in JSON responses
type PaginationResponse struct {
	CurrentPage  int  `json:"currentPage"`
	NextPage     *int `json:"nextPage"`
	PrevPage     *int `json:"prevPage"`
	TotalPage    int  `json:"totalPage"`
	TotalRecords int  `json:"totalRecords"`
}
