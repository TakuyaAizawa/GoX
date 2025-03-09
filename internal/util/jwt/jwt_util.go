package jwt

import (
	// "time"

	"github.com/google/uuid"
)

// JWTUtil JWTトークン操作のユーティリティ
type JWTUtil struct {
	secretKey     string
	accessExpiry  int
	refreshExpiry int
}

// NewJWTUtil 新しいJWTUtilを作成する
func NewJWTUtil(secretKey string, accessExpiry, refreshExpiry int) *JWTUtil {
	return &JWTUtil{
		secretKey:     secretKey,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateToken IDからアクセストークンを生成する
func (j *JWTUtil) GenerateToken(userID string) (string, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return "", err
	}
	return GenerateToken(id, "", "", AccessToken, j.secretKey, j.accessExpiry)
}

// GenerateTokenWithDetails ユーザー詳細を含むアクセストークンを生成する
func (j *JWTUtil) GenerateTokenWithDetails(userID, username, email string) (string, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return "", err
	}
	return GenerateToken(id, username, email, AccessToken, j.secretKey, j.accessExpiry)
}

// GenerateRefreshToken リフレッシュトークンを生成する
func (j *JWTUtil) GenerateRefreshToken(userID string) (string, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return "", err
	}
	return GenerateToken(id, "", "", RefreshToken, j.secretKey, j.refreshExpiry)
}

// ValidateAccessToken アクセストークンを検証する
func (j *JWTUtil) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := ValidateToken(tokenString, j.secretKey)
	if err != nil {
		return nil, err
	}

	// トークンタイプの検証
	if claims.Type != AccessToken {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// ValidateRefreshToken リフレッシュトークンを検証する
func (j *JWTUtil) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := ValidateToken(tokenString, j.secretKey)
	if err != nil {
		return nil, err
	}

	// トークンタイプの検証
	if claims.Type != RefreshToken {
		return nil, ErrInvalidTokenType
	}

	return claims, nil
}

// エラー定義
var (
	ErrInvalidTokenType = &TokenError{Message: "無効なトークンタイプです"}
)

// TokenError トークンエラーを表す構造体
type TokenError struct {
	Message string
}

// Error エラーメッセージを返す
func (e *TokenError) Error() string {
	return e.Message
}
