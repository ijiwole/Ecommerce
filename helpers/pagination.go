package helpers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page     int64
	PageSize int64
	Skip     int64
}

// GetPaginationParams extracts pagination parameters from query string
// Defaults: page=1, pageSize=10
func GetPaginationParams(c *gin.Context) PaginationParams {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 64)

	// Validate and set defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // Max page size limit
	}

	skip := (page - 1) * pageSize

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Skip:     skip,
	}
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination struct {
		Page       int64 `json:"page"`
		PageSize   int64 `json:"page_size"`
		Total      int64 `json:"total"`
		TotalPages int64 `json:"total_pages"`
	} `json:"pagination"`
}

// PaginatedSuccess sends a successful paginated response
func PaginatedSuccess(c *gin.Context, data interface{}, total int64, params PaginationParams) {
	totalPages := total / params.PageSize
	if total%params.PageSize > 0 {
		totalPages++
	}

	response := PaginatedResponse{
		Success: true,
		Data:    data,
	}
	response.Pagination.Page = params.Page
	response.Pagination.PageSize = params.PageSize
	response.Pagination.Total = total
	response.Pagination.TotalPages = totalPages

	c.JSON(200, response)
}
