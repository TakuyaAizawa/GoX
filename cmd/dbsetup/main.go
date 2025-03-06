package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/TakuyaAizawa/gox/pkg/database"
	"github.com/joho/godotenv"
)

func main() {
	// コマンドライン引数の解析
	var (
		envFile        = flag.String("env", ".env", "環境変数ファイルのパス")
		migrationsPath = flag.String("migrations", "migrations", "マイグレーションファイルのディレクトリパス")
		rollback       = flag.Bool("rollback", false, "最後のマイグレーションをロールバックする")
		version        = flag.Bool("version", false, "現在のマイグレーションバージョンを表示する")
	)
	flag.Parse()

	// 環境変数ファイルの読み込み
	loadEnvFile(*envFile)

	// データベースURLの取得
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("環境変数 DATABASE_URL が設定されていません")
	}

	// データベース設定
	config := &database.Config{
		URL:             dbURL,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
		MaxRetries:      5,
		RetryInterval:   5 * time.Second,
	}

	// データベース接続
	log.Println("データベースに接続しています...")
	db, err := database.NewPostgresDBWithConfig(config)
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer db.Close()

	// 接続テスト
	if err := database.TestConnection(db); err != nil {
		log.Fatalf("データベース接続テストに失敗しました: %v", err)
	}
	log.Println("データベース接続テストに成功しました")

	// マイグレーション設定
	migrationOptions := &database.MigrationOptions{
		MigrationsPath:  *migrationsPath,
		MigrationsTable: "schema_migrations",
		SchemaName:      "public",
	}

	// マイグレーションの実行
	if *rollback {
		// ロールバック
		log.Println("マイグレーションをロールバックしています...")
		if err := database.RollbackMigration(db, migrationOptions); err != nil {
			log.Fatalf("マイグレーションのロールバックに失敗しました: %v", err)
		}
		log.Println("マイグレーションのロールバックが完了しました")
	} else if *version {
		// バージョン表示
		log.Println("現在のマイグレーションバージョンを表示します")
		// TODO: バージョン表示の実装
		log.Println("注意: バージョン表示機能は未実装です")
	} else {
		// マイグレーション実行
		log.Println("マイグレーションを実行しています...")
		if err := database.RunMigrations(db, migrationOptions); err != nil {
			log.Fatalf("マイグレーションの実行に失敗しました: %v", err)
		}
		log.Println("マイグレーションが完了しました")
	}

	log.Println("データベースセットアップが正常に完了しました")
}

// loadEnvFile は環境変数ファイルを読み込みます
func loadEnvFile(envPath string) {
	// 絶対パスに変換
	absPath, err := filepath.Abs(envPath)
	if err != nil {
		log.Printf("警告: 環境変数ファイルパスの解決に失敗しました: %v", err)
		absPath = envPath
	}

	// ファイルの存在確認
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Printf("警告: 環境変数ファイルが見つかりません: %s", absPath)
		return
	}

	// 環境変数ファイルの読み込み
	if err := godotenv.Load(absPath); err != nil {
		log.Printf("警告: 環境変数ファイルの読み込みに失敗しました: %v", err)
		return
	}

	log.Printf("環境変数ファイルを読み込みました: %s", absPath)
} 