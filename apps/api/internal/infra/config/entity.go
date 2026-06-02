package infraconfig

// Config 是整个应用的配置。
type Config struct {
	Port     int            // 服务监听的端口号，比如 8080
	Database DatabaseConfig // 数据库的连接信息
	JWT      JWTConfig      // JWT 签名的密钥和有效期（明天用）
	Redis    RedisConfig    // Redis 的连接信息（第5天用）
	RabbitMQ RabbitMQConfig // 消息队列的连接信息（第10天用）
}

// DatabaseConfig 数据库连接需要的信息。
type DatabaseConfig struct {
	Host     string // 数据库所在机器的地址
	Port     int    // 数据库的端口，MySQL 默认 3306
	User     string // 数据库用户名
	Password string // 数据库密码
	Name     string // 数据库名字
}

// JWTConfig JWT 相关的配置。
type JWTConfig struct {
	Secret    string // 签名密钥
	AccessTTL string // token 有效期，比如 "15m" 表示 15 分钟
}

// RedisConfig Redis 连接信息。
type RedisConfig struct {
	Addr     string // Redis 地址，比如 "127.0.0.1:6379"
	Password string // Redis 密码
	DB       int    // 用第几个数据库（Redis 有 0-15 共 16 个）
}

// RabbitMQConfig 消息队列连接信息（第10天才用到）。
type RabbitMQConfig struct {
	URL string // RabbitMQ 的连接地址
}
