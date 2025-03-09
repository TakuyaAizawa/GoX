package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TakuyaAizawa/gox/internal/api/routes"
	"github.com/TakuyaAizawa/gox/internal/config"
	"github.com/TakuyaAizawa/gox/internal/repository/postgres"
	"github.com/TakuyaAizawa/gox/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

// @title GoX API
// @version 1.0
// @description GoXマイクロブログプラットフォームのAPI
// @termsOfService http://swagger.io/terms/

// @contact.name API サポート
// @contact.url http://www.yoursite.com/support
// @contact.email support@yoursite.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

func main() {
	// 設定のロード
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	// ロガーの初期化
	l, err := logger.NewLogger(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		log.Fatalf("ロガーの初期化に失敗しました: %v", err)
	}
	defer l.Sync()

	// データベース接続文字列の構築
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode)

	// コンテキストの作成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// データベース接続プールの設定
	dbConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		l.Fatal("データベース設定の解析に失敗しました", "error", err)
	}

	// プール接続の設定
	dbConfig.MaxConns = 10
	dbConfig.MinConns = 5
	dbConfig.MaxConnLifetime = 5 * time.Minute
	dbConfig.MaxConnIdleTime = 5 * time.Minute

	// データベース接続プールの作成
	db, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		l.Fatal("データベース接続に失敗しました", "error", err)
	}
	defer db.Close()

	// 接続テスト
	if err := db.Ping(ctx); err != nil {
		l.Fatal("データベース接続テストに失敗しました", "error", err)
	}
	l.Info("データベースに正常に接続しました")

	// リポジトリの初期化
	userRepo := postgres.NewUserRepository(db)
	postRepo := postgres.NewPostRepository(db)
	followRepo := postgres.NewFollowRepository(db)
	likeRepo := postgres.NewLikeRepository(db)
	notificationRepo := postgres.NewNotificationRepository(db)

	// ルーターのセットアップ
	router := routes.SetupRouter(
		cfg,
		l,
		userRepo,
		postRepo,
		followRepo,
		likeRepo,
		notificationRepo,
	)

	// HTTPサーバーの設定
	server := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// サーバーを非同期で起動
	go func() {
		l.Info("サーバーを起動中", "port", cfg.App.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatal("サーバーの起動に失敗しました", "error", err)
		}
	}()

	// グレースフルシャットダウンの設定
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	l.Info("サーバーをシャットダウンしています...")

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		l.Fatal("サーバーの強制シャットダウンが発生しました", "error", err)
	}

	l.Info("サーバーを終了します")
}
