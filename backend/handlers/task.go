package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/models"
	"prompt-vault/service"
)

// TaskHandler handles HTTP requests for tasks.
type TaskHandler struct {
	db  *gorm.DB
	svc *service.TaskService
}

// NewTaskHandler creates a new TaskHandler.
func NewTaskHandler(db *gorm.DB, svc *service.TaskService) *TaskHandler {
	return &TaskHandler{
		db:  db,
		svc: svc,
	}
}

// CreateTask godoc
// @Summary 创建任务
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body service.CreateTaskRequest true "任务参数"
// @Success 200 {object} map[string]interface{} "成功返回任务信息"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req service.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Validate task type
	validTypes := map[string]bool{
		models.TaskTypeBatchTest:  true,
		models.TaskTypeABTest:     true,
		models.TaskTypeEvalGen:    true,
		models.TaskTypeRegression: true,
		models.TaskTypeMultiTurn:  true,
	}
	if !validTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid task type"})
		return
	}

	task, err := h.svc.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    task.ToResponse(),
	})
}

// GetTask godoc
// @Summary 获取任务详情
// @Tags tasks
// @Produce json
// @Param id path int true "任务ID"
// @Success 200 {object} map[string]interface{} "成功返回任务信息"
// @Failure 400 {object} map[string]interface{} "无效的任务ID"
// @Failure 404 {object} map[string]interface{} "任务不存在"
// @Router /api/tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid task ID"})
		return
	}

	task, err := h.svc.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Task not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    task.ToResponse(),
	})
}

// ListTasks godoc
// @Summary 任务列表
// @Tags tasks
// @Produce json
// @Param limit query int false "每页数量，默认20"
// @Param offset query int false "偏移量，默认0"
// @Param status query string false "按状态筛选"
// @Success 200 {object} map[string]interface{} "成功返回任务列表"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status")

	var tasks []models.Task
	var total int64
	var err error

	if status != "" {
		tasks, total, err = h.svc.ListByStatus(status, limit, offset)
	} else {
		tasks, total, err = h.svc.List(limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	responses := make([]models.TaskResponse, len(tasks))
	for i, t := range tasks {
		responses[i] = t.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tasks": responses,
			"total": total,
		},
	})
}

// CancelTask godoc
// @Summary 取消任务
// @Tags tasks
// @Produce json
// @Param id path int true "任务ID"
// @Success 200 {object} map[string]interface{} "成功取消任务"
// @Failure 400 {object} map[string]interface{} "无效的任务ID"
// @Failure 404 {object} map[string]interface{} "任务不存在或无法取消"
// @Router /api/tasks/{id} [delete]
func (h *TaskHandler) CancelTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid task ID"})
		return
	}

	if err := h.svc.Cancel(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Task not found or cannot be cancelled"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Task cancelled successfully",
	})
}

// DeleteTask godoc
// @Summary 删除任务
// @Tags tasks
// @Produce json
// @Param id path int true "任务ID"
// @Success 200 {object} map[string]interface{} "成功删除任务"
// @Failure 400 {object} map[string]interface{} "无效的任务ID"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid task ID"})
		return
	}

	if err := h.svc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Task deleted successfully",
	})
}
