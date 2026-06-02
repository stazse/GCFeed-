package interfaceshttprouter

import (
	"database/sql"
	"log"

	infraconfig "GCFeed/internal/infra/config"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Register 是整个应用的"装配车间"。
// 所有组件（数据库、服务、处理器）在这里创建并连接起来。
func Register(g *gin.Engine, cfg *infraconfig.Config, db *sql.DB) error {
	// 用 GORM 包装 database/sql 的连接池
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: db, // 复用之前创建的连接池
	}), &gorm.Config{})
	if err != nil {
		return err
	}

	// _ = gormDB 表示"先留着，后面几天用"
	// 后面几天会在这里用 gormDB 自动创建表、注入到 handler
	_ = gormDB

	log.Println("database connection ready (gorm)")

	// ========== 注册路由 ==========

	// 健康检查接口
	g.GET("/health", HealthCheck)

	log.Println("routes registered")
	return nil
}

// HealthCheck 健康检查处理函数。
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "All is well",
	})
}
