package model

// Pagination represents standard domain pagination metadata
type Pagination struct {
	TotalRecords int  `json:"totalRecords"`
	CurrentPage  int  `json:"currentPage"`
	TotalPage    int  `json:"totalPage"`
	NextPage     *int `json:"nextPage"`
	PrevPage     *int `json:"prevPage"`
}

// BuildPagination constructs a Pagination model from total records, page, and page size
func BuildPagination(totalRecords, page, pageSize int) Pagination {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	totalPage := 0
	if totalRecords > 0 {
		totalPage = (totalRecords + pageSize - 1) / pageSize
	}

	var nextPage *int
	if page < totalPage {
		n := page + 1
		nextPage = &n
	}

	var prevPage *int
	if page > 1 && (totalPage == 0 || page <= totalPage+1) {
		p := page - 1
		prevPage = &p
	}

	return Pagination{
		TotalRecords: totalRecords,
		CurrentPage:  page,
		TotalPage:    totalPage,
		NextPage:     nextPage,
		PrevPage:     prevPage,
	}
}

// CalculateOffset returns the SQL offset for a given page and page size
func CalculateOffset(page, pageSize int) int {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return (page - 1) * pageSize
}
