package interfaceshttprouter

import (
	"database/sql"
	"log"

	applicationaccount "GCFeed/internal/application/account"
	applicationfeed "GCFeed/internal/application/feed"
	applicationvideo "GCFeed/internal/application/video"
	infracache "GCFeed/internal/infra/cache"
	infraconfig "GCFeed/internal/infra/config"
	infrajwt "GCFeed/internal/infra/jwt"
	infraaccount "GCFeed/internal/infra/persistence/account"
	infrafeed "GCFeed/internal/infra/persistence/feed"
	inframigration "GCFeed/internal/infra/persistence/migration"
	infravideo "GCFeed/internal/infra/persistence/video"
	interfaceshttpaccount "GCFeed/internal/interfaces/http/account"
	interfaceshttpfeed "GCFeed/internal/interfaces/http/feed"
	interfaceshttpmiddleware "GCFeed/internal/interfaces/http/middleware"
	interfaceshttpupload "GCFeed/internal/interfaces/http/upload"
	interfaceshttpvideo "GCFeed/internal/interfaces/http/video"

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

	// 视频模块装配
	videoRepo := infravideo.New(gormDB)
	videoService := applicationvideo.New(videoRepo)
	videoHandler := interfaceshttpvideo.New(videoService)

	// Feed 模块装配
	// Redis 是可选的：配置文件里有 Redis 地址才初始化
	var feedCache *infracache.FeedCache
	if cfg.Redis.Addr != "" {
		redisClient := infracache.NewRedisClient(cfg.Redis)
		feedCache = infracache.NewFeedCache(redisClient)
		log.Println("redis cache enabled")
	} else {
		log.Println("redis cache disabled (no addr configured)")
	}

	feedRepo := infrafeed.New(gormDB)
	// 用函数选项注入缓存
	var feedOpts []func(*applicationfeed.Service)
	if feedCache != nil {
		feedOpts = append(feedOpts, applicationfeed.WithFeedCache(feedCache))
	}
	feedService := applicationfeed.New(feedRepo, feedOpts...)
	feedHandler := interfaceshttpfeed.New(feedService)

	// 上传模块
	uploadHandler := interfaceshttpupload.New("./uploads")

	// 鉴权中间件
	authMiddleware := interfaceshttpmiddleware.NewJWTAuth(jwtManager)

	// 静态文件访问（让上传的文件可以通过 URL 直接访问）
	g.Static("/uploads", "./uploads")

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

	// 视频资源
	videos := api.Group("/videos")
	videos.POST("", authMiddleware, videoHandler.Create) // 发布视频（需登录）
	videos.GET("/:videoId", videoHandler.Get)            // 视频详情（公开）

	// 上传（需登录）
	uploadGroup := api.Group("/uploads", authMiddleware)
	uploadGroup.POST("", uploadHandler.Create)

	// 我的作品
	users.GET("/me/videos", authMiddleware, videoHandler.ListMine)

	// Feed 流
	api.GET("/feed-items", feedHandler.ListFeedItems)

	log.Println("routes registered")
	return nil
}

func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"message": "All is well"})
}
