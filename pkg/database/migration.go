package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrationOptions はマイグレーションの設定オプションを保持します
type MigrationOptions struct {
	// マイグレーションファイルのディレクトリパス
	MigrationsPath string
	
	// マイグレーションテーブル名
	MigrationsTable string
	
	// スキーマ名
	SchemaName string
}

// DefaultMigrationOptions はデフォルトのマイグレーション設定を返します
func DefaultMigrationOptions() *MigrationOptions {
	return &MigrationOptions{
		MigrationsPath:  "migrations",
		MigrationsTable: "schema_migrations",
		SchemaName:      "public",
	}
}

// RunMigrations はデータベースマイグレーションを実行します
func RunMigrations(db *PostgresDB, options *MigrationOptions) error {
	if db == nil {
		return errors.New("データベース接続がnilです")
	}
	
	if options == nil {
		options = DefaultMigrationOptions()
	}
	
	// マイグレーションパスの存在確認
	absPath, err := filepath.Abs(options.MigrationsPath)
	if err != nil {
		return fmt.Errorf("マイグレーションパスの解決に失敗しました: %w", err)
	}
	
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("マイグレーションディレクトリが存在しません: %s", absPath)
	}
	
	// Postgresドライバーの設定
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{
		MigrationsTable: options.MigrationsTable,
		SchemaName:      options.SchemaName,
	})
	if err != nil {
		return fmt.Errorf("マイグレーションドライバーの初期化に失敗しました: %w", err)
	}
	
	// マイグレーションの初期化
	// Windows対応のパス形式を使用
	cleanPath := filepath.ToSlash(absPath)
	sourceURL := "file://" + cleanPath
	log.Printf("マイグレーションソースURL: %s", sourceURL)
	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("マイグレーションの初期化に失敗しました: %w", err)
	}
	
	// マイグレーションの実行
	log.Println("データベースマイグレーションを実行しています...")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("マイグレーションの実行に失敗しました: %w", err)
	}
	
	// マイグレーションのバージョン確認
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("マイグレーションバージョンの取得に失敗しました: %w", err)
	}
	
	if dirty {
		log.Printf("警告: マイグレーションは「ダーティ」状態です (バージョン: %d)", version)
	} else if err == migrate.ErrNilVersion {
		log.Println("マイグレーションは実行されていません")
	} else {
		log.Printf("マイグレーションが正常に完了しました (現在のバージョン: %d)", version)
	}
	
	return nil
}

// RollbackMigration は最後のマイグレーションをロールバックします
func RollbackMigration(db *PostgresDB, options *MigrationOptions) error {
	if db == nil {
		return errors.New("データベース接続がnilです")
	}
	
	if options == nil {
		options = DefaultMigrationOptions()
	}
	
	// マイグレーションパスの存在確認
	absPath, err := filepath.Abs(options.MigrationsPath)
	if err != nil {
		return fmt.Errorf("マイグレーションパスの解決に失敗しました: %w", err)
	}
	
	// Postgresドライバーの設定
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{
		MigrationsTable: options.MigrationsTable,
		SchemaName:      options.SchemaName,
	})
	if err != nil {
		return fmt.Errorf("マイグレーションドライバーの初期化に失敗しました: %w", err)
	}
	
	// マイグレーションの初期化
	// Windows対応のパス形式を使用
	cleanPath := filepath.ToSlash(absPath)
	sourceURL := "file://" + cleanPath
	log.Printf("マイグレーションソースURL: %s", sourceURL)
	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("マイグレーションの初期化に失敗しました: %w", err)
	}
	
	// 1つ前のバージョンにロールバック
	log.Println("最後のマイグレーションをロールバックしています...")
	if err := m.Steps(-1); err != nil {
		return fmt.Errorf("マイグレーションのロールバックに失敗しました: %w", err)
	}
	
	log.Println("マイグレーションのロールバックが完了しました")
	return nil
}