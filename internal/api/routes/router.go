package routes

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/TakuyaAizawa/gox/internal/api/middleware"
	"github.com/TakuyaAizawa/gox/internal/config"
	"github.com/TakuyaAizawa/gox/pkg/logger"
)

// APIルートを設定する
func SetupRouter(cfg *config.Config, log logger.Logger) *gin.Engine {
	// プロダクションモードの場合はデバッグモードを無効化
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

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
		// 認証不要のエンドポイント
		// 認証ルートグループ
		v1.Group("/auth")
		// {
			// TODO: 認証ハンドラーを追加
			// auth.POST("/register", handlers.Register)
			// auth.POST("/login", handlers.Login)
			// auth.POST("/refresh", handlers.RefreshToken)
		// }

		// 認証が必要なエンドポイント
		secured := v1.Group("")
		secured.Use(middleware.Auth(cfg.JWT.Secret))
		{
			// ユーザー関連
			// ユーザールートグループ
			secured.Group("/users")
			// {
				// TODO: ユーザーハンドラーを追加
				// users.GET("", handlers.GetUsers)
				// users.GET("/:id", handlers.GetUser)
				// users.PUT("/:id", handlers.UpdateUser)
				// users.DELETE("/:id", handlers.DeleteUser)
			// }

			// 投稿関連
			// 投稿ルートグループ
			secured.Group("/posts")
			// {
				// TODO: 投稿ハンドラーを追加
				// posts.POST("", handlers.CreatePost)
				// posts.GET("", handlers.GetPosts)
				// posts.GET("/:id", handlers.GetPost)
				// posts.PUT("/:id", handlers.UpdatePost)
				// posts.DELETE("/:id", handlers.DeletePost)
			// }

			// タイムライン関連
			// タイムラインルートグループ
			secured.Group("/timeline")
			// {
				// TODO: タイムラインハンドラーを追加
				// timeline.GET("", handlers.GetTimeline)
			// }

			// フォロー関連
			// フォロールートグループ
			secured.Group("/follows")
			// {
				// TODO: フォローハンドラーを追加
				// follows.POST("/:id", handlers.FollowUser)
				// follows.DELETE("/:id", handlers.UnfollowUser)
				// follows.GET("/followers", handlers.GetFollowers)
				// follows.GET("/following", handlers.GetFollowing)
			// }

			// 通知関連
			// 通知ルートグループ
			secured.Group("/notifications")
			// {
				// TODO: 通知ハンドラーを追加
				// notifications.GET("", handlers.GetNotifications)
				// notifications.PUT("/:id/read", handlers.MarkNotificationAsRead)
			// }
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
		// TODO: 本番環境ではSPAのindex.htmlを返すよう設定
		c.JSON(http.StatusNotFound, gin.H{
			"error": "見つかりません",
		})
	})

	return r
} 