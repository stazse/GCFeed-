package infrahttpgin

import (
	"fmt"
	"net/http"

	infraconfig "GCFeed/internal/infra/config"

	"github.com/gin-gonic/gin"
)

// Init 创建一个 Gin 引擎实例。
func Init() *gin.Engine {
	g := gin.New()

	// Recovery 中间件：当程序 panic（崩溃）时，自动恢复并记录日志
	g.Use(gin.Recovery())

	// Logger 中间件：每个请求都会自动打印一行日志
	g.Use(gin.Logger())

	return g
}

// Run 启动 HTTP 服务，开始监听端口。
func Run(cfg *infraconfig.Config, g *gin.Engine) error {
	addr := fmt.Sprintf(":%d", cfg.Port)
	return http.ListenAndServe(addr, g)
}
