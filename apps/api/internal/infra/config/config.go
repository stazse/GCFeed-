package infraconfig

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfig 从 yaml 文件中读取配置。
func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// 设置默认端口
	if cfg.Port == 0 {
		cfg.Port = 8080
	}

	return &cfg, nil
}
