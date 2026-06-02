package infradatabase

import (
	"database/sql"
	"fmt"

	infraconfig "GCFeed/internal/infra/config"

	_ "github.com/go-sql-driver/mysql"
)

// New 创建一个数据库连接池。
// ⚠️ 关键：如果目标数据库不存在，会自动创建。
func New(cfg infraconfig.DatabaseConfig) (*sql.DB, error) {
	// ========== 第 1 步：先连接 MySQL 服务器（不指定数据库） ==========
	// DSN 中不写数据库名，只连到 MySQL 服务器本身
	serverDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)

	initDB, err := sql.Open("mysql", serverDSN)
	if err != nil {
		return nil, fmt.Errorf("open init connection: %w", err)
	}
	defer initDB.Close()

	// 测试 MySQL 服务器是否可达
	if err := initDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql server: %w", err)
	}

	// ========== 第 2 步：创建数据库（如果不存在） ==========
	createSQL := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.Name,
	)
	if _, err := initDB.Exec(createSQL); err != nil {
		return nil, fmt.Errorf("create database %s: %w", cfg.Name, err)
	}

	// ========== 第 3 步：连接到目标数据库 ==========
	targetDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	db, err := sql.Open("mysql", targetDSN)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// 验证目标数据库连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25) // 最多同时打开 25 个连接
	db.SetMaxIdleConns(10) // 空闲时保留 10 个连接

	return db, nil
}
