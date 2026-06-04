package infraconfig

// Config 是整个应用的配置。
type Config struct {
	Port     int            `yaml:"port"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Redis    RedisConfig    `yaml:"redis"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
}

// DatabaseConfig 数据库连接需要的信息。
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

// JWTConfig JWT 相关的配置。
type JWTConfig struct {
	Secret    string `yaml:"secret"`
	AccessTTL string `yaml:"access_ttl"`
}

// RedisConfig Redis 连接信息。
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// RabbitMQConfig 消息队列连接信息（第10天才用到）。
type RabbitMQConfig struct {
	URL string `yaml:"url"`
}
