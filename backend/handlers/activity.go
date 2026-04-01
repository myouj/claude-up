package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/middleware"
	"prompt-vault/models"
)

type ActivityHandler struct {
	db *gorm.DB
}

func NewActivityHandler(db *gorm.DB) *ActivityHandler {
	return &ActivityHandler{db: db}
}

// Log records an activity asynchronously. In test mode (TESTING=1) it runs synchronously
// to avoid race conditions with DB cleanup between tests.
func (h *ActivityHandler) Log(entityType string, entityID uint, action, details string) {
	fn := func() {
		defer func() {
			if r := recover(); r != nil {
				middleware.Error("activity log panic recovered", map[string]interface{}{
				"error": r,
			})
			}
		}()
		entry := models.ActivityLog{
			EntityType: entityType,
			EntityID:   entityID,
			Action:     action,
			Details:    details,
		}
		h.db.Create(&entry)
	}

	if os.Getenv("TESTING") == "1" {
		fn()
		return
	}
	go fn()
}

func (h *ActivityHandler) List(c *gin.Context) {
	var logs []models.ActivityLog
	countQuery := h.db.Model(&models.ActivityLog{})
	query := h.db.Order("created_at DESC")

	if entityType := c.Query("entity_type"); entityType != "" {
		query = query.Where("entity_type = ?", entityType)
		countQuery = countQuery.Where("entity_type = ?", entityType)
	}
	if entityID := c.Query("entity_id"); entityID != "" {
		query = query.Where("entity_id = ?", entityID)
		countQuery = countQuery.Where("entity_id = ?", entityID)
	}
	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
		countQuery = countQuery.Where("action = ?", action)
	}

	offset, _, limit, _, meta := middleware.ParsePagination(c, countQuery, query)
	query.Offset(offset).Limit(limit).Find(&logs)

	var responses []models.ActivityLogResponse
	for _, l := range logs {
		responses = append(responses, models.ActivityLogResponse{
			ID:         l.ID,
			EntityType: l.EntityType,
			EntityID:   l.EntityID,
			Action:     l.Action,
			UserID:     l.UserID,
			Details:    l.Details,
			CreatedAt:  l.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Success: true,
		Data:    responses,
		Meta:    meta,
	})
}
