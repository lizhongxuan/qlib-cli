package main

import (
	"log"
	"net/http"

	"qlib-backend/config"
	"qlib-backend/internal/api/routes"
	"qlib-backend/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	if err := services.InitDatabase(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 自动迁移数据库表结构
	if err := services.AutoMigrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 设置Gin模式
	gin.SetMode(cfg.App.Mode)

	// 创建路由
	r := gin.Default()

	// 配置CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 设置路由
	routes.SetupRoutes(r)

	// 启动服务器
	log.Printf("Server starting on port %s", cfg.App.Port)
	if err := http.ListenAndServe(":"+cfg.App.Port, r); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}