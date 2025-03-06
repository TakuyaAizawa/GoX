package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWT認証のためのミドルウェア
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorization ヘッダーの取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorizationヘッダーが必要です"})
			c.Abort()
			return
		}

		// Bearer トークンの形式を確認
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorizationヘッダーの形式はBearer {token}である必要があります"})
			c.Abort()
			return
		}

		// JWT トークンの検証
		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// アルゴリズムの検証
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("予期しない署名方式です: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("無効なトークンです: %v", err)})
			c.Abort()
			return
		}

		// トークンのクレームを取得
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// ユーザーIDを取得
			userID, ok := claims["sub"].(string)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "無効なトークンペイロードです"})
				c.Abort()
				return
			}

			// ユーザーIDをコンテキストに設定
			c.Set("user_id", userID)
			
			// ユーザー情報を必要に応じて設定
			if username, ok := claims["username"].(string); ok {
				c.Set("username", username)
			}
			
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "無効なトークンです"})
			c.Abort()
			return
		}
	}
} 