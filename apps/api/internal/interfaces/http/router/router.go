package interfaceshttprouter

import (
	"context"
	"database/sql"
	"log"

	applicationaccount "GCFeed/internal/application/account"
	applicationfeed "GCFeed/internal/application/feed"
	applicationinteraction "GCFeed/internal/application/interaction"
	applicationrelation "GCFeed/internal/application/relation"
	applicationvideo "GCFeed/internal/application/video"
	infracache "GCFeed/internal/infra/cache"
	infraconfig "GCFeed/internal/infra/config"
	infrajwt "GCFeed/internal/infra/jwt"
	infraaccount "GCFeed/internal/infra/persistence/account"
	infrafeed "GCFeed/internal/infra/persistence/feed"
	infrainteraction "GCFeed/internal/infra/persistence/interaction"
	inframigration "GCFeed/internal/infra/persistence/migration"
	infrarelation "GCFeed/internal/infra/persistence/relation"
	infravideo "GCFeed/internal/infra/persistence/video"
	interfaceshttpaccount "GCFeed/internal/interfaces/http/account"
	interfaceshttpfeed "GCFeed/internal/interfaces/http/feed"
	interfaceshttpinteraction "GCFeed/internal/interfaces/http/interaction"
	interfaceshttpmiddleware "GCFeed/internal/interfaces/http/middleware"
	interfaceshttprelation "GCFeed/internal/interfaces/http/relation"
	interfaceshttpupload "GCFeed/internal/interfaces/http/upload"
	interfaceshttpvideo "GCFeed/internal/interfaces/http/video"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Register(g *gin.Engine, cfg *infraconfig.Config, db *sql.DB) error {
	// 用 GORM 包装连接池
	gormDB, err := gorm.Open(mysql.New(mysql.Config{Conn: db}), &gorm.Config{
		TranslateError:                           true,
		DisableForeignKeyConstraintWhenMigrating: true, // 避免 GORM 误将唯一索引当 FK 删除
	})
	if err != nil {
		return err
	}

	// AutoMigrate：根据 Go 结构体自动创建/更新数据库表
	if err := inframigration.AutoMigrate(gormDB); err != nil {
		log.Printf("auto-migrate warning: %v (continuing anyway)", err)
		// 不阻塞启动：表已存在时继续运行
	} else {
		log.Println("database migrated")
	}

	// 创建 JWT 管理器
	jwtManager, err := infrajwt.NewManager(&cfg.JWT)
	if err != nil {
		return err
	}

	// 视频模块装配
	videoRepo := infravideo.New(gormDB)
	videoService := applicationvideo.New(videoRepo)
	videoHandler := interfaceshttpvideo.New(videoService)

	// Redis 是可选的：配置文件里有 Redis 地址才初始化
	var feedCache *infracache.FeedCache
	if cfg.Redis.Addr != "" {
		redisClient := infracache.NewRedisClient(cfg.Redis)
		// 验证 Redis 是否真正可达，不可达则不启用缓存
		if err := infracache.Ping(context.Background(), redisClient); err != nil {
			log.Printf("redis ping failed (%v), cache disabled", err)
		} else {
			feedCache = infracache.NewFeedCache(redisClient)
			log.Println("redis cache enabled")
		}
	} else {
		log.Println("redis cache disabled (no addr configured)")
	}

	// --- Feed 模块装配 ---
	feedRepo := infrafeed.New(gormDB)
	feedOptions := []func(*applicationfeed.Service){}
	if feedCache != nil {
		feedOptions = append(feedOptions,
			applicationfeed.WithFeedCache(feedCache),
			applicationfeed.WithHotProvider(feedCache), // feedCache 同时实现 HotFeedProvider
		)
	}
	// feedCache == nil 时 hotProvider 也为 nil，getHotFeed 会降级为 timeline
	feedService := applicationfeed.New(feedRepo, feedOptions...)
	feedHandler := interfaceshttpfeed.New(feedService)

	// --- 互动模块装配 ---
	interactionRepo := infrainteraction.New(gormDB)
	interactionOptions := []func(*applicationinteraction.Service){}
	if feedCache != nil {
		interactionOptions = append(interactionOptions,
			applicationinteraction.WithHotScoreRecorder(feedCache))
	}
	interactionService := applicationinteraction.New(interactionRepo, interactionOptions...)
	interactionHandler := interfaceshttpinteraction.New(interactionService)

	// --- 关注关系装配 ---
	relationRepo := infrarelation.New(gormDB)
	relationService := applicationrelation.New(relationRepo)
	relationHandler := interfaceshttprelation.New(relationService)

	// 上传模块
	uploadHandler := interfaceshttpupload.New("./uploads")

	// 鉴权中间件
	authMiddleware := interfaceshttpmiddleware.NewJWTAuth(jwtManager)

	// 静态文件访问
	g.Static("/uploads", "./uploads")

	// 用户模块装配
	accountRepo := infraaccount.New(gormDB)
	accountService := applicationaccount.New(accountRepo, jwtManager)
	accountHandler := interfaceshttpaccount.New(accountService)

	// ========== 注册路由 ==========
	g.GET("/health", HealthCheck)

	api := g.Group("/api")

	// 用户资源
	users := api.Group("/users")
	users.POST("", accountHandler.Register)

	// 会话资源
	sessions := api.Group("/sessions")
	sessions.POST("", accountHandler.Login)

	// 视频资源
	videos := api.Group("/videos")
	videos.POST("", authMiddleware, videoHandler.Create)
	videos.GET("/:videoId", videoHandler.Get)

	// 上传（需登录）
	uploadGroup := api.Group("/uploads", authMiddleware)
	uploadGroup.POST("", uploadHandler.Create)

	// 我的作品
	users.GET("/me/videos", authMiddleware, videoHandler.ListMine)

	// Feed 流
	api.GET("/feed-items", feedHandler.ListFeedItems)

	// 互动路由
	videos.PUT("/:videoId/like", authMiddleware, interactionHandler.Like)
	videos.DELETE("/:videoId/like", authMiddleware, interactionHandler.Unlike)
	videos.PUT("/:videoId/favorite", authMiddleware, interactionHandler.Favorite)
	videos.DELETE("/:videoId/favorite", authMiddleware, interactionHandler.Unfavorite)
	videos.POST("/:videoId/comments", authMiddleware, interactionHandler.CreateComment)
	videos.GET("/:videoId/comments", interactionHandler.ListComments)
	api.DELETE("/comments/:commentId", authMiddleware, interactionHandler.DeleteComment)

	// 关注关系路由
	users.PUT("/me/following/:targetUserId", authMiddleware, relationHandler.Follow)
	users.DELETE("/me/following/:targetUserId", authMiddleware, relationHandler.Unfollow)
	users.GET("/me/following", authMiddleware, relationHandler.ListFollowing)
	users.GET("/me/followers", authMiddleware, relationHandler.ListFollowers)

	log.Println("routes registered")
	return nil
}

func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"message": "All is well"})
}
