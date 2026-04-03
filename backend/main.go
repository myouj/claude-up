package main

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prompt-vault/handlers"
	"prompt-vault/middleware"
	"prompt-vault/models"
	"prompt-vault/service"
	"prompt-vault/worker"
)

var db *gorm.DB

func main() {
	// 加载 .env 环境变量
	_ = godotenv.Load()

	// 初始化数据库
	var err error
	db, err = gorm.Open(sqlite.Open("prompt-vault.db"), &gorm.Config{})
	if err != nil {
		middleware.Fatal("failed to connect database", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 自动迁移
	err = db.AutoMigrate(
		&models.Prompt{},
		&models.PromptVersion{},
		&models.TestRecord{},
		&models.Skill{},
		&models.Agent{},
		&models.Translation{},
		&models.ActivityLog{},
		&models.Setting{},
		&models.AICallLog{},
		&models.Task{},
		&models.EvalSet{},
	)
	if err != nil {
		middleware.Fatal("failed to migrate database", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 初始化示例数据
	initSampleData()

	// 初始化日志系统
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "" {
		switch strings.ToUpper(logLevel) {
		case "DEBUG":
			middleware.SetLevel(middleware.DEBUG)
		case "INFO":
			middleware.SetLevel(middleware.INFO)
		case "WARN":
			middleware.SetLevel(middleware.WARN)
		case "ERROR":
			middleware.SetLevel(middleware.ERROR)
		}
	}

	middleware.Info("server starting", map[string]interface{}{
		"version": "1.0",
		"log_level": logLevel,
	})

	// 初始化处理器
	activityHandler := handlers.NewActivityHandler(db)
	promptHandler := handlers.NewPromptHandler(db, activityHandler)
	versionHandler := handlers.NewVersionHandler(db)
	testHandler := handlers.NewTestHandler(db, activityHandler)
	skillHandler := handlers.NewSkillHandler(db, activityHandler)
	agentHandler := handlers.NewAgentHandler(db, activityHandler)
	translateHandler := handlers.NewTranslateHandler(db)
	settingHandler := handlers.NewSettingHandler(db)
	taskService := service.NewTaskService(db)
	taskHandler := handlers.NewTaskHandler(db, taskService)
	scoringHandler := handlers.NewScoringHandler(db)
	evalHandler := handlers.NewEvalHandler(db)

	// 初始化 Worker Pool
	workerPool := worker.NewPool(worker.DefaultWorkerConfig(), db)
	workerPool.Start()

	// 路由配置
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 全局限流器放在 trace 之前以便统计 IP
	rl := newRateLimiter(100, time.Minute)

	// 全局中间件：恢复(带日志) -> TraceId -> 请求日志 -> AI调用日志
	r.Use(middleware.RecoveryLoggerMiddleware())
	r.Use(middleware.TraceMiddleware())
	r.Use(middleware.RequestLoggerMiddleware())
	r.Use(middleware.AICallLogMiddleware(db))

	// 限流中间件（使用已在上面初始化的 rl）
	r.Use(rateLimitMiddleware(rl))

	// CORS 中间件
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsEnv != "" {
		allowedOrigins = strings.Split(allowedOriginsEnv, ",")
	}
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" && len(allowedOrigins) > 0 {
			for _, allowed := range allowedOrigins {
				if strings.TrimSpace(allowed) == origin {
					c.Header("Access-Control-Allow-Origin", origin)
					c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
					break
				}
			}
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API 路由
	api := r.Group("/api")
	{
		// 提示词 CRUD
		api.GET("/prompts", promptHandler.List)
		api.POST("/prompts", promptHandler.Create)
		api.GET("/prompts/:id", promptHandler.Get)
		api.PUT("/prompts/:id", promptHandler.Update)
		api.DELETE("/prompts/:id", promptHandler.Delete)
		api.POST("/prompts/:id/favorite", promptHandler.ToggleFavorite)
		api.GET("/prompts/categories", promptHandler.ListCategories)
		api.POST("/prompts/:id/clone", promptHandler.Clone)
		api.GET("/prompts/export", promptHandler.Export)
		api.POST("/prompts/import", promptHandler.Import)

		// 版本管理
		api.GET("/prompts/:id/versions", versionHandler.List)
		api.POST("/prompts/:id/versions", versionHandler.Create)
		api.GET("/versions/:id", versionHandler.Get)

		// 测试与优化
		api.POST("/prompts/:id/test", testHandler.Test)
		api.POST("/prompts/:id/optimize", testHandler.Optimize)
		api.GET("/prompts/:id/tests", testHandler.List)
		api.GET("/prompts/:id/tests/compare", testHandler.Compare)
		api.GET("/prompts/:id/test-compare", testHandler.Compare)
		api.GET("/prompts/:id/analytics", testHandler.Analytics)
		api.GET("/models", testHandler.ListModels)

		// Skills CRUD
		api.GET("/skills", skillHandler.List)
		api.POST("/skills", skillHandler.Create)
		api.GET("/skills/:id", skillHandler.Get)
		api.PUT("/skills/:id", skillHandler.Update)
		api.DELETE("/skills/:id", skillHandler.Delete)
		api.GET("/skills/categories", skillHandler.ListCategories)
		api.POST("/skills/:id/clone", skillHandler.Clone)
		api.GET("/skills/export", skillHandler.Export)
		api.POST("/skills/import", skillHandler.Import)

		// Agents CRUD
		api.GET("/agents", agentHandler.List)
		api.POST("/agents", agentHandler.Create)
		api.GET("/agents/:id", agentHandler.Get)
		api.PUT("/agents/:id", agentHandler.Update)
		api.DELETE("/agents/:id", agentHandler.Delete)
		api.GET("/agents/categories", agentHandler.ListCategories)
		api.POST("/agents/:id/clone", agentHandler.Clone)
		api.GET("/agents/export", agentHandler.Export)
		api.POST("/agents/import", agentHandler.Import)

		// 翻译
		api.POST("/translate", translateHandler.Translate)
		api.POST("/translate/:type/:id", translateHandler.TranslateEntity)

		// 活动日志
		api.GET("/activity-logs", activityHandler.List)

		// 设置
		api.GET("/settings", settingHandler.List)
		api.GET("/settings/:key", settingHandler.Get)
		api.PUT("/settings/:key", settingHandler.Set)
		api.DELETE("/settings/:key", settingHandler.Delete)

		// 任务
		api.GET("/tasks", taskHandler.ListTasks)
		api.POST("/tasks", taskHandler.CreateTask)
		api.GET("/tasks/:id", taskHandler.GetTask)
		api.DELETE("/tasks/:id", taskHandler.CancelTask)

		// 评分
		api.GET("/prompts/:id/score", scoringHandler.ScorePrompt)
		api.POST("/prompts/score-batch", scoringHandler.ScoreBatch)
		api.GET("/scoring/weights", scoringHandler.GetWeights)

		// 评测集
		api.GET("/prompts/:id/eval-sets", evalHandler.ListEvalSets)
		api.POST("/prompts/:id/eval-sets", evalHandler.CreateEvalSet)
		api.GET("/eval-sets/:id", evalHandler.GetEvalSet)
		api.PUT("/eval-sets/:id", evalHandler.UpdateEvalSet)
		api.DELETE("/eval-sets/:id", evalHandler.DeleteEvalSet)
		api.POST("/prompts/:id/eval-sets/generate", evalHandler.GenerateAutoEvalSet)
		api.POST("/prompts/:id/eval-sets/:eval_id/run", evalHandler.RunEval)
	}

	// 统计 API
	api.GET("/stats", func(c *gin.Context) {
		var promptCount int64
		var skillCount int64
		var agentCount int64
		db.Model(&models.Prompt{}).Count(&promptCount)
		db.Model(&models.Skill{}).Count(&skillCount)
		db.Model(&models.Agent{}).Count(&agentCount)

		c.JSON(200, gin.H{
			"success": true,
			"data": gin.H{
				"prompts": promptCount,
				"skills":  skillCount,
				"agents":  agentCount,
			},
		})
	})

	// 全量导出
	api.GET("/export", func(c *gin.Context) {
		var prompts []models.Prompt
		var skills []models.Skill
		var agents []models.Agent
		db.Order("updated_at DESC").Find(&prompts)
		db.Order("updated_at DESC").Find(&skills)
		db.Order("updated_at DESC").Find(&agents)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": models.ExportPayload{
				Version:    "1.0",
				ExportedAt: time.Now().Format("2006-01-02 15:04:05"),
				Prompts:    prompts,
				Skills:     skills,
				Agents:     agents,
			},
		})
	})

	middleware.Info("server listening", map[string]interface{}{
		"addr": ":8080",
	})
	r.Run(":8080")
}

// ----- Rate Limiter -----

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	// Start background cleanup goroutine to prevent memory leak.
	go rl.gc()
	return rl
}

// gc periodically cleans up expired entries to prevent unbounded memory growth.
func (rl *rateLimiter) gc() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)
		for ip, times := range rl.requests {
			valid := make([]time.Time, 0, len(times))
			for _, t := range times {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = valid
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Filter old requests
	valid := rl.requests[ip][:0]
	for _, t := range rl.requests[ip] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	rl.requests[ip] = valid

	if len(rl.requests[ip]) >= rl.limit {
		return false
	}
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}

func rateLimitMiddleware(rl *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !rl.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func initSampleData() {
	// 检查是否已有数据
	var skillCount int64
	var agentCount int64
	db.Model(&models.Skill{}).Count(&skillCount)
	db.Model(&models.Agent{}).Count(&agentCount)

	// 初始化 Skills
	if skillCount == 0 {
		skills := []models.Skill{
			{
				Name:        "/commit",
				Description: "Generate a semantic git commit message based on staged changes",
				Content:     "You are a git commit message generator. Analyze the staged changes and write a clear, concise commit message following conventional commits format.",
				Category:    "git",
				Source:      "builtin",
			},
			{
				Name:        "/review-pr",
				Description: "Review pull requests and provide constructive feedback",
				Content:     "You are a code reviewer. Analyze the pull request and provide constructive feedback on code quality, potential bugs, security issues, and improvement suggestions.",
				Category:    "code",
				Source:      "builtin",
			},
			{
				Name:        "/explain-code",
				Description: "Explain code in detail",
				Content:     "You are a code documentation expert. Explain the provided code in detail, covering its purpose, logic, and key components.",
				Category:    "docs",
				Source:      "builtin",
			},
		}
		for _, s := range skills {
			db.Create(&s)
		}
	}

	// 初始化 Agents
	if agentCount == 0 {
		agents := []models.Agent{
			{
				Name:         "code-reviewer",
				Role:         "Code Reviewer",
				Content:      "You are an expert code reviewer. Analyze code thoroughly for bugs, security vulnerabilities, performance issues, and adherence to best practices. Provide clear, actionable feedback.",
				Capabilities: "Static analysis, Security review, Performance optimization, Best practices",
				Category:     "development",
				Source:       "builtin",
			},
			{
				Name:         "security-expert",
				Role:         "Security Expert",
				Content:      "You are a cybersecurity expert. Analyze code and systems for security vulnerabilities, suggest mitigations, and recommend security best practices.",
				Capabilities: "Vulnerability assessment, Security auditing, Penetration testing, Compliance",
				Category:     "security",
				Source:       "builtin",
			},
			{
				Name:         "documentation-writer",
				Role:         "Technical Writer",
				Content:      "You are a technical documentation expert. Write clear, comprehensive documentation including README files, API docs, and user guides.",
				Capabilities: "Technical writing, API documentation, README files, User guides",
				Category:     "docs",
				Source:       "builtin",
			},
		}
		for _, a := range agents {
			db.Create(&a)
		}
	}
}
