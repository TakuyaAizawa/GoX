package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
)

// CORSを処理するミドルウェアを返す
func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// デバッグ用ログ出力
		log.Printf("Received request from origin: %s", origin)
		log.Printf("Allowed origins: %v", allowedOrigins)

		// オリジンが許可されているかチェック
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin || allowedOrigin == "*" {
				allowed = true
				break
			}
		}

		// CORSヘッダーを設定
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			// 開発環境ではすべてのオリジンを許可（本番環境では使用しないでください）
			c.Header("Access-Control-Allow-Origin", origin)
			log.Printf("Warning: Origin %s is not in the allowed list, but allowing it anyway for development", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		// プリフライトリクエストを処理
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
