package middleware

import (
	// "net/http"
	"strings"

	"github.com/TakuyaAizawa/gox/internal/util/jwt"
	"github.com/TakuyaAizawa/gox/internal/util/response"
	"github.com/TakuyaAizawa/gox/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// JWT認証のためのミドルウェア
func Auth(jwtUtil *jwt.JWTUtil, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorization ヘッダーの取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "認証が必要です")
			c.Abort()
			return
		}

		// Bearer トークンの形式を確認
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "認証形式が無効です")
			c.Abort()
			return
		}

		// JWT トークンの検証
		tokenString := parts[1]
		claims, err := jwtUtil.ValidateAccessToken(tokenString)
		if err != nil {
			log.Info("トークン検証に失敗しました", "error", err)
			response.Unauthorized(c, "無効なトークンです")
			c.Abort()
			return
		}

		// ユーザーIDを文字列からUUIDに変換
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			log.Error("ユーザーIDのパースに失敗しました", "error", err, "userID", claims.UserID)
			response.InternalServerError(c, "認証情報の処理に失敗しました")
			c.Abort()
			return
		}

		// ユーザーIDをコンテキストに設定（UUID型で設定）
		c.Set("userID", userID)

		// その他のユーザー情報を必要に応じて設定
		if claims.Username != "" {
			c.Set("username", claims.Username)
		}
		if claims.Email != "" {
			c.Set("email", claims.Email)
		}

		c.Next()
	}
}
