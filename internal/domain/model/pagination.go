package model

import "math"

// Pagination represents internal domain pagination data & helpers
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// Offset returns the SQL query offset
func (p Pagination) Offset() int {
	if p.Page <= 0 {
		return 0
	}
	return (p.Page - 1) * p.Limit
}

// TotalPages calculates total page count
func (p Pagination) TotalPages(totalRecords int) int {
	if p.Limit <= 0 {
		return 0
	}
	return int(math.Ceil(float64(totalRecords) / float64(p.Limit)))
}
