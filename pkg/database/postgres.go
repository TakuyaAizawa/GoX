package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// DB はデータベース接続のインターフェースを定義します
type DB interface {
	// 基本的なデータベース操作メソッド
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	
	// トランザクション関連
	Begin() (*sql.Tx, error)
	
	// 接続管理
	Close() error
	Ping() error
	
	// コンテキスト対応メソッド
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	PingContext(ctx context.Context) error
}

// PostgresDB は*sql.DBをラップしてDB interfaceを実装します
type PostgresDB struct {
	*sql.DB
	config *Config
}

// NewPostgresDB は新しいPostgreSQLデータベース接続を作成します
func NewPostgresDB(dbURL string) (*PostgresDB, error) {
	return NewPostgresDBWithConfig(&Config{
		URL:             dbURL,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		MaxRetries:      5,
		RetryInterval:   5 * time.Second,
	})
}

// NewPostgresDBWithConfig は設定を指定してPostgreSQLデータベース接続を作成します
func NewPostgresDBWithConfig(config *Config) (*PostgresDB, error) {
	if config == nil {
		return nil, errors.New("データベース設定がnilです")
	}
	
	if config.URL == "" {
		return nil, errors.New("データベースURLが空です")
	}
	
	var db *sql.DB
	var err error
	
	// リトライロジックを実装
	for i := 0; i <= config.MaxRetries; i++ {
		db, err = sql.Open("postgres", config.URL)
		if err != nil {
			log.Printf("データベース接続の初期化に失敗しました (試行 %d/%d): %v", 
				i+1, config.MaxRetries+1, err)
			
			if i < config.MaxRetries {
				time.Sleep(config.RetryInterval)
				continue
			}
			return nil, fmt.Errorf("データベース接続の初期化に失敗しました: %w", err)
		}
		
		// コネクションプールの設定
		db.SetMaxOpenConns(config.MaxOpenConns)
		db.SetMaxIdleConns(config.MaxIdleConns)
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
		
		// 接続テスト
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = db.PingContext(ctx)
		cancel()
		
		if err != nil {
			log.Printf("データベース接続テストに失敗しました (試行 %d/%d): %v", 
				i+1, config.MaxRetries+1, err)
			
			if i < config.MaxRetries {
				time.Sleep(config.RetryInterval)
				continue
			}
			return nil, fmt.Errorf("データベース接続テストに失敗しました: %w", err)
		}
		
		// 接続成功
		log.Printf("データベースに正常に接続しました (試行 %d/%d)", i+1, config.MaxRetries+1)
		break
	}
	
	return &PostgresDB{
		DB:     db,
		config: config,
	}, nil
}

// TestConnection はデータベース接続が正常に機能しているかテストします
func TestConnection(db DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// 簡単なクエリを実行
	var one int
	err := db.QueryRowContext(ctx, "SELECT 1").Scan(&one)
	if err != nil {
		return fmt.Errorf("接続テストクエリの実行に失敗しました: %w", err)
	}
	
	if one != 1 {
		return fmt.Errorf("予期しない結果: 期待値=1, 実際の値=%d", one)
	}
	
	return nil
}

// Close はデータベース接続を閉じます
func (pdb *PostgresDB) Close() error {
	log.Println("データベース接続を閉じています...")
	return pdb.DB.Close()
} 