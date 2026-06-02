package main

import (
	"log"

	infraconfig "GCFeed/internal/infra/config"
	infradatabase "GCFeed/internal/infra/database"
	infrahttpgin "GCFeed/internal/infra/httpgin"
	interfaceshttprouter "GCFeed/internal/interfaces/http/router"
)

func main() {
	// ========== 第 1 步：读取配置 ==========
	cfg, err := infraconfig.LoadConfig("./configs/config.yaml")
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}
	log.Printf("config loaded, port=%d", cfg.Port)

	// ========== 第 2 步：连接数据库（自动创建库） ==========
	db, err := infradatabase.New(cfg.Database)
	if err != nil {
		log.Fatalf("init database failed: %v", err)
	}
	log.Println("database connected")

	// ========== 第 3 步：创建 Gin 引擎 ==========
	g := infrahttpgin.Init()
	log.Println("gin engine initialized")

	// ========== 第 4 步：注册路由 ==========
	if err := interfaceshttprouter.Register(g, cfg, db); err != nil {
		log.Fatalf("register routes failed: %v", err)
	}
	log.Println("router registered")

	// ========== 第 5 步：启动服务 ==========
	log.Printf("server is starting on http://127.0.0.1:%d", cfg.Port)
	if err := infrahttpgin.Run(cfg, g); err != nil {
		log.Fatalf("run server failed: %v", err)
	}
}
