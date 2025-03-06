package database

import "time"

// Config はデータベース接続の設定を保持する構造体です
type Config struct {
	// 接続文字列
	URL string

	// コネクションプールの設定
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// 接続試行の設定
	MaxRetries    int
	RetryInterval time.Duration
}

// DefaultConfig はデフォルトのデータベース設定を返します
func DefaultConfig() *Config {
	return &Config{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		MaxRetries:      5,
		RetryInterval:   5 * time.Second,
	}
} 