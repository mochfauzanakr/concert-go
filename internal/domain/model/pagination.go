package model

import (
	"strings"
	"time"
)

// BaseListParams holds standard query parameters for paginated list endpoints
type BaseListParams struct {
	Page      int        `json:"page"`
	PageSize  int        `json:"pageSize"`
	Search    string     `json:"search"`
	SortBy    string     `json:"sortBy"`
	SortOrder string     `json:"sortOrder"`
	StartDate *time.Time `json:"startDate,omitempty"`
	EndDate   *time.Time `json:"endDate,omitempty"`
}

// Normalize sanitizes pagination, search, and sort direction defaults
func (p *BaseListParams) Normalize(maxPageSize int) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if maxPageSize > 0 && p.PageSize > maxPageSize {
		p.PageSize = maxPageSize
	}
	p.Search = strings.TrimSpace(p.Search)
	p.SortBy = strings.ToLower(strings.TrimSpace(p.SortBy))
	order := strings.ToUpper(strings.TrimSpace(p.SortOrder))
	if order != "ASC" && order != "DESC" {
		order = "DESC"
	}
	p.SortOrder = order
}

// ResolveSortColumn validates sortBy against an allowed map (queryKey -> dbColumn) with a safe fallback
func ResolveSortColumn(sortBy string, allowed map[string]string, defaultCol string) string {
	key := strings.ToLower(strings.TrimSpace(sortBy))
	if col, ok := allowed[key]; ok {
		return col
	}
	return defaultCol
}

// ResolveSortOrder returns "ASC" or "DESC" safely
func ResolveSortOrder(sortOrder string) string {
	if strings.EqualFold(strings.TrimSpace(sortOrder), "ASC") {
		return "ASC"
	}
	return "DESC"
}

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
