package interfaceshttprouter

import (
	"database/sql"
	"log"

	applicationaccount "GCFeed/internal/application/account"
	infraconfig "GCFeed/internal/infra/config"
	infrajwt "GCFeed/internal/infra/jwt"
	infraaccount "GCFeed/internal/infra/persistence/account"
	inframigration "GCFeed/internal/infra/persistence/migration"
	interfaceshttpaccount "GCFeed/internal/interfaces/http/account"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Register(g *gin.Engine, cfg *infraconfig.Config, db *sql.DB) error {
	// 用 GORM 包装连接池
	gormDB, err := gorm.Open(mysql.New(mysql.Config{Conn: db}), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		return err
	}

	// AutoMigrate：根据 Go 结构体自动创建/更新数据库表
	// 这是 GORM 提供的一个非常方便的功能：你不用手写 CREATE TABLE
	if err := inframigration.AutoMigrate(gormDB); err != nil {
		return err
	}
	log.Println("database migrated")

	// 创建 JWT 管理器
	jwtManager, err := infrajwt.NewManager(&cfg.JWT)
	if err != nil {
		return err
	}

	// 装配：Repository → Service → Handler
	accountRepo := infraaccount.New(gormDB)
	accountService := applicationaccount.New(accountRepo, jwtManager)
	accountHandler := interfaceshttpaccount.New(accountService)

	// 注册路由
	g.GET("/health", HealthCheck)

	api := g.Group("/api")

	// 用户资源
	users := api.Group("/users")
	users.POST("", accountHandler.Register) // POST /api/users → 注册

	// 会话资源（登录=创建会话）
	sessions := api.Group("/sessions")
	sessions.POST("", accountHandler.Login) // POST /api/sessions → 登录

	log.Println("routes registered")
	return nil
}

func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"message": "All is well"})
}
