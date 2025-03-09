package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// アプリケーション設定を表す構造体
type Config struct {
	App       AppConfig
	DB        DBConfig
	Redis     RedisConfig
	JWT       JWTConfig
	CORS      CORSConfig
	Log       LogConfig
	RateLimit RateLimitConfig
	Storage   StorageConfig
}

// アプリケーション固有の設定を保持する構造体
type AppConfig struct {
	Env  string
	Port string
	Name string
	URL  string
}

// データベース接続設定を保持する構造体
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// Redis接続設定を保持する構造体
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// JWT認証設定を保持する構造体
type JWTConfig struct {
	Secret            string
	ExpirationHours   int
	RefreshExpiration int
}

// CORS設定を保持する構造体
type CORSConfig struct {
	AllowedOrigins []string
}

// ログ設定を保持する構造体
type LogConfig struct {
	Level  string
	Format string
}

// レート制限設定を保持する構造体
type RateLimitConfig struct {
	Requests int
	Duration time.Duration
}

// ストレージ設定を保持する構造体
type StorageConfig struct {
	Provider string
	BaseDir  string
	BaseURL  string
}

// 環境変数と.envファイルから設定を読み込む
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// デフォルト値の設定
	setDefaults()

	// .envファイルの読み込み (なければ環境変数から)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
		}
	}

	var config Config
	config.App = AppConfig{
		Env:  viper.GetString("app.env"),
		Port: viper.GetString("app.port"),
		Name: viper.GetString("app.name"),
		URL:  viper.GetString("app.url"),
	}

	config.DB = DBConfig{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		User:     viper.GetString("db.user"),
		Password: viper.GetString("db.password"),
		Name:     viper.GetString("db.name"),
		SSLMode:  viper.GetString("db.sslmode"),
	}

	config.Redis = RedisConfig{
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetString("redis.port"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	}

	config.JWT = JWTConfig{
		Secret:            viper.GetString("jwt.secret"),
		ExpirationHours:   viper.GetInt("jwt.expiration_hours"),
		RefreshExpiration: viper.GetInt("jwt.refresh_expiration_days"),
	}

	config.CORS = CORSConfig{
		AllowedOrigins: viper.GetStringSlice("cors.allowed_origins"),
	}

	config.Log = LogConfig{
		Level:  viper.GetString("log.level"),
		Format: viper.GetString("log.format"),
	}

	config.RateLimit = RateLimitConfig{
		Requests: viper.GetInt("rate_limit.requests"),
		Duration: time.Duration(viper.GetInt("rate_limit.duration")) * time.Second,
	}

	config.Storage = StorageConfig{
		Provider: viper.GetString("storage.provider"),
		BaseDir:  viper.GetString("storage.base_dir"),
		BaseURL:  viper.GetString("storage.base_url"),
	}

	return &config, nil
}

// 設定のデフォルト値を設定する
func setDefaults() {
	// アプリケーションのデフォルト値
	viper.SetDefault("app.env", "development")
	viper.SetDefault("app.port", "8080")
	viper.SetDefault("app.name", "GoX")
	viper.SetDefault("app.url", "http://localhost:8080")

	// データベースのデフォルト値
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", "5432")
	viper.SetDefault("db.user", "postgres")
	viper.SetDefault("db.password", "postgres")
	viper.SetDefault("db.name", "gox")
	viper.SetDefault("db.sslmode", "disable")

	// Redisのデフォルト値
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// JWTのデフォルト値
	viper.SetDefault("jwt.expiration_hours", 24)
	viper.SetDefault("jwt.refresh_expiration_days", 7)

	// CORSのデフォルト値
	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:3000", "http://localhost:8080"})

	// ログのデフォルト値
	viper.SetDefault("log.level", "debug")
	viper.SetDefault("log.format", "json")

	// レート制限のデフォルト値
	viper.SetDefault("rate_limit.requests", 100)
	viper.SetDefault("rate_limit.duration", 60)

	// ストレージのデフォルト値
	viper.SetDefault("storage.provider", "local")
	viper.SetDefault("storage.base_dir", "./uploads")
	viper.SetDefault("storage.base_url", "http://localhost:8080/media")
}
