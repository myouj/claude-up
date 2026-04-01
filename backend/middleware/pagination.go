package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/models"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// ParsePagination extracts page and limit from query params.
// Returns offset, page, limit, total count, and pagination metadata.
// IMPORTANT: countQuery must NOT have Order() applied — Count() + Order() is broken in GORM+SQLite.
func ParsePagination(c *gin.Context, countQuery *gorm.DB, orderedQuery *gorm.DB) (offset, page, limit int, total int64, meta models.PaginationMeta) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, _ = strconv.Atoi(pageStr)
	limit, _ = strconv.Atoi(limitStr)

	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	offset = (page - 1) * limit
	countQuery.Count(&total)

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return offset, page, limit, total, models.PaginationMeta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}

// CalcTotalPages computes total pages from total count and limit.
func CalcTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := int(total) / limit
	if int(total)%limit != 0 {
		pages++
	}
	return pages
}
