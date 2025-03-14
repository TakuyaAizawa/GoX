package routes

import (
	"net/http"
	"strings"

	"github.com/TakuyaAizawa/gox/internal/api/handlers"
	"github.com/TakuyaAizawa/gox/internal/api/middleware"
	"github.com/TakuyaAizawa/gox/internal/config"
	coreinterfaces "github.com/TakuyaAizawa/gox/internal/interfaces"
	repointerfaces "github.com/TakuyaAizawa/gox/internal/repository/interfaces"
	"github.com/TakuyaAizawa/gox/internal/service"
	"github.com/TakuyaAizawa/gox/internal/storage"
	"github.com/TakuyaAizawa/gox/internal/util/jwt"
	"github.com/TakuyaAizawa/gox/pkg/logger"
	"github.com/gin-gonic/gin"
)

// SetupRouter APIルートを設定する
func SetupRouter(
	cfg *config.Config,
	log logger.Logger,
	userRepo repointerfaces.UserRepository,
	postRepo repointerfaces.PostRepository,
	followRepo repointerfaces.FollowRepository,
	likeRepo repointerfaces.LikeRepository,
	notificationRepo repointerfaces.NotificationRepository,
) *gin.Engine {
	// プロダクションモードの場合はデバッグモードを無効化
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// JWTユーティリティの作成
	jwtUtil := jwt.NewJWTUtil(cfg.JWT.Secret, cfg.JWT.ExpirationHours, cfg.JWT.RefreshExpiration)

	r := gin.New()

	// ミドルウェアの設定
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))
	r.Use(middleware.RateLimit(cfg.RateLimit.Requests, cfg.RateLimit.Duration))

	// メディアファイルの静的配信
	r.Static("/media", cfg.Storage.BaseDir)

	// ヘルスチェックエンドポイント
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API v1 ルート
	v1 := r.Group("/api/v1")

	// ストレージプロバイダーの作成
	var storageProvider coreinterfaces.StorageProvider
	if cfg.Storage.Provider == "local" {
		storageProvider = storage.NewLocalStorage(cfg.Storage.BaseDir, cfg.Storage.BaseURL, log)
	} else {
		log.Warn("ストレージプロバイダー設定が無効です。ローカルストレージを使用します", "provider", cfg.Storage.Provider)
		storageProvider = storage.NewLocalStorage(cfg.Storage.BaseDir, cfg.Storage.BaseURL, log)
	}

	// ハンドラーの作成
	authHandler := handlers.NewAuthHandler(userRepo, log, jwtUtil)
	wsHandler := handlers.NewWebSocketHandler(log)

	// 通知サービス
	notificationService := service.NewNotificationService(
		notificationRepo,
		userRepo,
		postRepo,
		wsHandler.GetNotificationHub(),
		log,
	)

	// ユーザーハンドラー
	userHandler := handlers.NewUserHandler(
		userRepo,
		followRepo,
		postRepo,
		notificationService,
		storageProvider,
		log,
	)

	// 投稿ハンドラー
	postHandler := handlers.NewPostHandler(
		postRepo,
		userRepo,
		likeRepo,
		notificationRepo,
		notificationService,
		log,
	)

	// タイムラインハンドラー
	timelineHandler := handlers.NewTimelineHandler(
		postRepo,
		userRepo,
		followRepo,
		likeRepo,
		log,
	)

	// 通知ハンドラー
	notificationHandler := handlers.NewNotificationHandler(
		notificationRepo,
		userRepo,
		postRepo,
		log,
	)

	// 認証エンドポイント
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", authHandler.Logout)
	}

	// 認証が必要なエンドポイント
	secured := v1.Group("")
	secured.Use(middleware.Auth(jwtUtil, log))
	{
		// ユーザー関連
		users := secured.Group("/users")
		{
			// ユーザープロフィール
			users.GET("/:username", userHandler.GetUserProfile)
			users.PUT("/me", userHandler.UpdateProfile)

			// プロフィール画像アップロード
			users.POST("/me/avatar", userHandler.UploadAvatar)
			users.POST("/me/banner", userHandler.UploadBanner)

			// フォロー関連
			users.POST("/:username/follow", userHandler.FollowUser)
			users.DELETE("/:username/follow", userHandler.UnfollowUser)
			users.GET("/:username/followers", userHandler.GetFollowers)
			users.GET("/:username/following", userHandler.GetFollowing)

			// ユーザーの投稿
			users.GET("/:username/posts", userHandler.GetUserPosts)
		}

		// 投稿関連
		posts := secured.Group("/posts")
		{
			posts.POST("", postHandler.CreatePost)
			posts.GET("/:id", postHandler.GetPost)
			posts.DELETE("/:id", postHandler.DeletePost)

			// 返信
			posts.GET("/:id/replies", postHandler.GetPostReplies)

			// いいね
			posts.POST("/:id/like", postHandler.LikePost)
			posts.DELETE("/:id/like", postHandler.UnlikePost)

			// TODO: リポスト機能
			// posts.POST("/:id/repost", postHandler.RepostPost)
			// posts.DELETE("/:id/repost", postHandler.CancelRepost)
		}

		// タイムライン関連
		timeline := secured.Group("/timeline")
		{
			timeline.GET("/home", timelineHandler.GetHomeTimeline)
			timeline.GET("/explore", timelineHandler.GetExploreTimeline)
		}

		// 通知エンドポイント
		notifications := secured.Group("/notifications")
		{
			notifications.GET("", notificationHandler.GetNotifications)
			notifications.GET("/unread", notificationHandler.GetUnreadCount)
			notifications.PUT("/read", notificationHandler.MarkAsRead)
		}
	}

	// WebSocketエンドポイント
	v1.GET("/ws", middleware.Auth(jwtUtil, log), wsHandler.HandleWSConnection)

	// 404ハンドラー
	r.NoRoute(func(c *gin.Context) {
		// APIルートのみ処理
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "見つかりません",
			})
			return
		}

		// SPAのフロントエンドへのフォールバック
		c.JSON(http.StatusNotFound, gin.H{
			"error": "見つかりません",
		})
	})

	return r
}
