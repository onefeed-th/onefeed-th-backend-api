package config

import (
	"context"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	RestServer restServer `mapstructure:"restServer"`
	Postgres   postgres   `mapstructure:"postgres"`
	Redis      redis      `mapstructure:"redis"`
}

type restServer struct {
	Port int `mapstructure:"port"`
}

type postgres struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Dbname   string `mapstructure:"dbname"`
	Pool     postgresPool `mapstructure:"pool"`
}

type postgresPool struct {
	MaxConns        int32 `mapstructure:"maxConns"`
	MinConns        int32 `mapstructure:"minConns"`
	MaxConnLifetime int   `mapstructure:"maxConnLifetime"` // in minutes
	MaxConnIdleTime int   `mapstructure:"maxConnIdleTime"` // in minutes
	HealthCheckPeriod int `mapstructure:"healthCheckPeriod"` // in minutes
	ConnectTimeout  int   `mapstructure:"connectTimeout"`    // in seconds
}

type redis struct {
	Host     string    `mapstructure:"host"`
	Port     int       `mapstructure:"port"`
	Password string    `mapstructure:"password"`
	Pool     redisPool `mapstructure:"pool"`
}

type redisPool struct {
	PoolSize        int `mapstructure:"poolSize"`
	MinIdleConns    int `mapstructure:"minIdleConns"`
	MaxIdleConns    int `mapstructure:"maxIdleConns"`
	PoolTimeout     int `mapstructure:"poolTimeout"`     // in seconds
	IdleTimeout     int `mapstructure:"idleTimeout"`     // in minutes
	MaxConnAge      int `mapstructure:"maxConnAge"`      // in minutes
	DialTimeout     int `mapstructure:"dialTimeout"`     // in seconds
	ReadTimeout     int `mapstructure:"readTimeout"`     // in seconds
	WriteTimeout    int `mapstructure:"writeTimeout"`    // in seconds
	MaxRetries      int `mapstructure:"maxRetries"`
	MinRetryBackoff int `mapstructure:"minRetryBackoff"` // in milliseconds
	MaxRetryBackoff int `mapstructure:"maxRetryBackoff"` // in milliseconds
}

var config *Config

func Init(ctx context.Context, configPath string) error {
	var err error
	config, err = LoadConfig(ctx, configPath)
	return err
}

func LoadConfig(ctx context.Context, configPath string) (*Config, error) {
	// Set up environment variable support
	// setEnvKeyReplacer allows nested config keys (like restServer.port) 
	// to be overridden by environment variables (REST_SERVER_PORT)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // Enable automatic environment variable binding
	
	// Set reasonable defaults
	setDefaults()

	// Read configuration file if provided
	if configPath != "" {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			// Don't fail if config file is missing - env vars and defaults will be used
			// This allows for container deployments with only env vars
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("restServer.port", 8080)

	// Database connection defaults (not credentials)
	viper.SetDefault("postgres.host", "localhost")
	viper.SetDefault("postgres.port", 5432)
	// Note: No defaults for user, password, dbname - these must be provided

	// PostgreSQL Pool defaults
	viper.SetDefault("postgres.pool.maxConns", 25)
	viper.SetDefault("postgres.pool.minConns", 5)
	viper.SetDefault("postgres.pool.maxConnLifetime", 60)      // 1 hour
	viper.SetDefault("postgres.pool.maxConnIdleTime", 30)      // 30 minutes
	viper.SetDefault("postgres.pool.healthCheckPeriod", 1)     // 1 minute
	viper.SetDefault("postgres.pool.connectTimeout", 5)        // 5 seconds

	// Redis connection defaults (not password)
	viper.SetDefault("redis.host", "localhost") 
	viper.SetDefault("redis.port", 6379)
	// Note: No default for password - it must be provided if required

	// Redis Pool defaults
	viper.SetDefault("redis.pool.poolSize", 15)
	viper.SetDefault("redis.pool.minIdleConns", 5)
	viper.SetDefault("redis.pool.maxIdleConns", 10)
	viper.SetDefault("redis.pool.poolTimeout", 4)              // 4 seconds
	viper.SetDefault("redis.pool.idleTimeout", 5)              // 5 minutes
	viper.SetDefault("redis.pool.maxConnAge", 30)              // 30 minutes
	viper.SetDefault("redis.pool.dialTimeout", 5)              // 5 seconds
	viper.SetDefault("redis.pool.readTimeout", 3)              // 3 seconds
	viper.SetDefault("redis.pool.writeTimeout", 3)             // 3 seconds
	viper.SetDefault("redis.pool.maxRetries", 2)
	viper.SetDefault("redis.pool.minRetryBackoff", 8)          // 8 milliseconds
	viper.SetDefault("redis.pool.maxRetryBackoff", 512)        // 512 milliseconds
}

func GetConfig() *Config {
	return config
}

// ResolveConfigFromFile exists for backward compatibility
func ResolveConfigFromFile(ctx context.Context, configPath string) (*Config, error) {
	return LoadConfig(ctx, configPath)
}
