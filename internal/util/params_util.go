package util

import (
	"strconv"
	"strings"
	"time"

	"concert-go/internal/domain/model"

	"github.com/gin-gonic/gin"
)

// ParseQueryInt extracts a positive integer query parameter with a default fallback
func ParseQueryInt(c *gin.Context, key string, defaultVal int) int {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return defaultVal
	}
	if v, err := strconv.Atoi(raw); err == nil && v > 0 {
		return v
	}
	return defaultVal
}

// ParseQueryTime parses an ISO-8601 (RFC3339) or YYYY-MM-DD date string from query parameters
func ParseQueryTime(c *gin.Context, key string, endOfDay bool) *time.Time {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02", raw); err == nil {
		if endOfDay {
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
		return &t
	}
	return nil
}

// ParseBaseListParams extracts standard pagination, search, sorting, and date range parameters
func ParseBaseListParams(c *gin.Context) model.BaseListParams {
	page := ParseQueryInt(c, "page", 1)
	pageSize := ParseQueryInt(c, "pageSize", 10)
	if c.Query("pageSize") == "" && c.Query("limit") != "" {
		pageSize = ParseQueryInt(c, "limit", 10)
	}

	search := strings.TrimSpace(c.Query("search"))
	if search == "" {
		search = strings.TrimSpace(c.Query("searchQuery"))
	}

	params := model.BaseListParams{
		Page:      page,
		PageSize:  pageSize,
		Search:    search,
		SortBy:    c.Query("sortBy"),
		SortOrder: c.Query("sortOrder"),
		StartDate: ParseQueryTime(c, "startDate", false),
		EndDate:   ParseQueryTime(c, "endDate", true),
	}
	params.Normalize(20)
	return params
}
