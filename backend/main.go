package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prompt-vault/handlers"
	"prompt-vault/models"
)

var db *gorm.DB

func main() {
	// 加载 .env 环境变量
	_ = godotenv.Load()

	// 初始化数据库
	var err error
	db, err = gorm.Open(sqlite.Open("prompt-vault.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// 自动迁移
	err = db.AutoMigrate(
		&models.Prompt{},
		&models.PromptVersion{},
		&models.TestRecord{},
		&models.Skill{},
		&models.Agent{},
		&models.Translation{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 初始化示例数据
	initSampleData()

	// 初始化处理器
	promptHandler := handlers.NewPromptHandler(db)
	versionHandler := handlers.NewVersionHandler(db)
	testHandler := handlers.NewTestHandler(db)
	skillHandler := handlers.NewSkillHandler(db)
	agentHandler := handlers.NewAgentHandler(db)
	translateHandler := handlers.NewTranslateHandler(db)

	// 路由配置
	r := gin.Default()

	// CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
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

		// 版本管理
		api.GET("/prompts/:id/versions", versionHandler.List)
		api.POST("/prompts/:id/versions", versionHandler.Create)
		api.GET("/versions/:id", versionHandler.Get)

		// 测试与优化
		api.POST("/prompts/:id/test", testHandler.Test)
		api.POST("/prompts/:id/optimize", testHandler.Optimize)
		api.GET("/prompts/:id/tests", testHandler.List)

		// Skills CRUD
		api.GET("/skills", skillHandler.List)
		api.POST("/skills", skillHandler.Create)
		api.GET("/skills/:id", skillHandler.Get)
		api.PUT("/skills/:id", skillHandler.Update)
		api.DELETE("/skills/:id", skillHandler.Delete)

		// Agents CRUD
		api.GET("/agents", agentHandler.List)
		api.POST("/agents", agentHandler.Create)
		api.GET("/agents/:id", agentHandler.Get)
		api.PUT("/agents/:id", agentHandler.Update)
		api.DELETE("/agents/:id", agentHandler.Delete)

		// 翻译
		api.POST("/translate", translateHandler.Translate)
		api.POST("/translate/:type/:id", translateHandler.TranslateEntity)
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

	log.Println("Server starting on :8080...")
	r.Run(":8080")
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
