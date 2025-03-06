package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTトークンの種類を定義
type TokenType string

const (
	// APIアクセスに使用するアクセストークン
	AccessToken TokenType = "access"
	
	// 新しいアクセストークンを取得するためのリフレッシュトークン
	RefreshToken TokenType = "refresh"
)

// JWTクレームを表す構造体
type Claims struct {
	UserID   string    `json:"sub"`
	Username string    `json:"username,omitempty"`
	Email    string    `json:"email,omitempty"`
	Type     TokenType `json:"type"`
	jwt.RegisteredClaims
}

// 新しいJWTトークンを生成する
func GenerateToken(userID uuid.UUID, username, email string, tokenType TokenType, secret string, expirationHours int) (string, error) {
	// 有効期限の設定
	expirationTime := time.Now().Add(time.Duration(expirationHours) * time.Hour)
	
	// クレームの作成
	claims := &Claims{
		UserID:   userID.String(),
		Username: username,
		Email:    email,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "gox-api",
		},
	}
	
	// トークンの作成
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// トークンの署名
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("トークンの署名に失敗しました: %w", err)
	}
	
	return tokenString, nil
}

// JWTトークンを検証し、クレームを返す
func ValidateToken(tokenString, secret string) (*Claims, error) {
	// JWTトークンの解析と検証
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 署名アルゴリズムの検証
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("予期しない署名方式です: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("トークンの解析に失敗しました: %w", err)
	}
	
	// クレームの取得
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, fmt.Errorf("無効なトークンです")
}

// トークンクレームからユーザーIDを抽出する
func GetUserIDFromToken(claims *Claims) (uuid.UUID, error) {
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("トークン内のユーザーIDが無効です: %w", err)
	}
	return userID, nil
} 