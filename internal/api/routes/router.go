package routes

import (
	"net/http"
	"strings"

	"github.com/TakuyaAizawa/gox/internal/api/handlers"
	"github.com/TakuyaAizawa/gox/internal/api/middleware"
	"github.com/TakuyaAizawa/gox/internal/config"
	"github.com/TakuyaAizawa/gox/internal/repository/interfaces"
	"github.com/TakuyaAizawa/gox/internal/util/jwt"
	"github.com/TakuyaAizawa/gox/pkg/logger"
	"github.com/gin-gonic/gin"
)

// SetupRouter APIルートを設定する
func SetupRouter(
	cfg *config.Config,
	log logger.Logger,
	userRepo interfaces.UserRepository,
	postRepo interfaces.PostRepository,
	followRepo interfaces.FollowRepository,
	likeRepo interfaces.LikeRepository,
	notificationRepo interfaces.NotificationRepository,
) *gin.Engine {
	// プロダクションモードの場合はデバッグモードを無効化
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// JWTユーティリティの作成
	jwtUtil := jwt.NewJWTUtil(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)

	r := gin.New()

	// ミドルウェアの設定
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))
	r.Use(middleware.RateLimit(cfg.RateLimit.Requests, cfg.RateLimit.Duration))

	// ヘルスチェックエンドポイント
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API v1 ルート
	v1 := r.Group("/api/v1")
	{
		// ハンドラーの作成
		authHandler := handlers.NewAuthHandler(userRepo, log, jwtUtil)
		userHandler := handlers.NewUserHandler(userRepo, followRepo, postRepo, log)
		postHandler := handlers.NewPostHandler(postRepo, userRepo, likeRepo, notificationRepo, log)
		timelineHandler := handlers.NewTimelineHandler(postRepo, userRepo, followRepo, likeRepo, log)
		notificationHandler := handlers.NewNotificationHandler(notificationRepo, userRepo, postRepo, log)

		// 認証不要のエンドポイント
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

				// TODO: プロフィール画像アップロード
				// users.POST("/me/avatar", userHandler.UploadAvatar)
				// users.POST("/me/banner", userHandler.UploadBanner)

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

			// 通知関連
			notifications := secured.Group("/notifications")
			{
				notifications.GET("", notificationHandler.GetNotifications)
				notifications.GET("/unread", notificationHandler.GetUnreadCount)
				notifications.PUT("/read", notificationHandler.MarkAsRead)
			}
		}
	}

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
