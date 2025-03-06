package testing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestDB はテストで使用するデータベース接続を表します
type TestDB struct {
	Pool *pgxpool.Pool
	URL  string
}

// NewTestDB は新しいテストデータベース接続を作成します
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	// テスト用のデータベースURLを取得
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Fatal("TEST_DATABASE_URL environment variable is not set")
	}

	// マイグレーションの実行
	if err := runMigrations(t, dbURL); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// コンテキストの作成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// データベース接続プールの設定
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		t.Fatalf("Failed to parse database URL: %v", err)
	}

	// プール接続の設定
	config.MaxConns = 5
	config.MinConns = 1
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	// データベース接続プールの作成
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create connection pool: %v", err)
	}

	// 接続テスト
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	return &TestDB{
		Pool: pool,
		URL:  dbURL,
	}
}

// runMigrations はマイグレーションを実行します
func runMigrations(t *testing.T, dbURL string) error {
	t.Helper()

	// プロジェクトのルートディレクトリを見つける（GoXディレクトリ）
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %v", err)
	}
	t.Logf("Current working directory: %s", wd)

	// "internal/repository/postgres/testing" から "migrations" へのパスを構築
	dir := wd
	for {
		if filepath.Base(dir) == "GoX" {
			break
		}
		dir = filepath.Dir(dir)
		if dir == "." || dir == "/" {
			return fmt.Errorf("could not find project root (GoX directory)")
		}
	}
	t.Logf("Project root directory: %s", dir)

	migrationsPath := filepath.Join(dir, "migrations")
	// ディレクトリの存在確認
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory not found at %s", migrationsPath)
	}
	t.Logf("Migrations directory: %s", migrationsPath)

	migrationsURL := fmt.Sprintf("file://%s", migrationsPath)
	t.Logf("Migrations URL: %s", migrationsURL)

	// マイグレーションインスタンスの作成
	m, err := migrate.New(migrationsURL, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}
	defer m.Close()

	// マイグレーションの実行
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	t.Log("Migrations completed successfully")
	return nil
}

// Close はテストデータベース接続を閉じます
func (db *TestDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// CleanupTable は指定されたテーブルのデータをクリーンアップします
func (db *TestDB) CleanupTable(t *testing.T, tableName string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName)
	_, err := db.Pool.Exec(ctx, query)
	if err != nil {
		t.Errorf("Failed to cleanup table %s: %v", tableName, err)
	}
}

// CleanupAllTables はすべてのテストテーブルをクリーンアップします
func (db *TestDB) CleanupAllTables(t *testing.T) {
	t.Helper()

	// 外部キー制約を考慮して、正しい順序でクリーンアップ
	tables := []string{
		"notifications",
		"likes",
		"posts",
		"follows",
		"users",
	}

	for _, table := range tables {
		db.CleanupTable(t, table)
	}
}

// WithTransaction はトランザクション内でテストを実行します
func (db *TestDB) WithTransaction(t *testing.T, fn func(tx pgx.Tx)) {
	t.Helper()

	ctx := context.Background()
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r)
		}
	}()

	fn(tx)

	err = tx.Rollback(ctx)
	if err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}
}
