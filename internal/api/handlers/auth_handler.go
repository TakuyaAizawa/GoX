package handlers

import (
	"net/http"
	"time"

	"github.com/TakuyaAizawa/gox/internal/domain/models"
	"github.com/TakuyaAizawa/gox/internal/repository/interfaces"
	"github.com/TakuyaAizawa/gox/internal/util/jwt"
	"github.com/TakuyaAizawa/gox/internal/util/response"
	"github.com/TakuyaAizawa/gox/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler 認証関連のハンドラーを管理する構造体
type AuthHandler struct {
	userRepo interfaces.UserRepository
	log      logger.Logger
	jwtUtil  *jwt.JWTUtil
}

// NewAuthHandler 新しい認証ハンドラーを作成する
func NewAuthHandler(userRepo interfaces.UserRepository, log logger.Logger, jwtUtil *jwt.JWTUtil) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		log:      log,
		jwtUtil:  jwtUtil,
	}
}

// RegisterRequest ユーザー登録リクエストの構造体
type RegisterRequest struct {
	Username    string `json:"username" binding:"required,alphanum,min=3,max=30"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=50"`
}

// Register ユーザー登録ハンドラー
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// ユーザー名とメールアドレスの使用可否をチェック
	usernameAvailable, err := h.userRepo.IsUsernameAvailable(c, req.Username)
	if err != nil {
		h.log.Error("ユーザー名の確認中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "ユーザー名の確認中にエラーが発生しました")
		return
	}
	if !usernameAvailable {
		response.BadRequest(c, "このユーザー名は既に使用されています", nil)
		return
	}

	emailAvailable, err := h.userRepo.IsEmailAvailable(c, req.Email)
	if err != nil {
		h.log.Error("メールアドレスの確認中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "メールアドレスの確認中にエラーが発生しました")
		return
	}
	if !emailAvailable {
		response.BadRequest(c, "このメールアドレスは既に使用されています", nil)
		return
	}

	// パスワードをハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.log.Error("パスワードのハッシュ化中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "パスワードのハッシュ化中にエラーが発生しました")
		return
	}

	// 新しいユーザーを作成
	now := time.Now().UTC()
	user := &models.User{
		ID:        uuid.New(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Name:      req.DisplayName,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.userRepo.Create(c, user); err != nil {
		h.log.Error("ユーザーの作成中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "ユーザーの作成中にエラーが発生しました")
		return
	}

	// JWTトークンを生成
	token, err := h.jwtUtil.GenerateToken(user.ID.String())
	if err != nil {
		h.log.Error("トークンの生成中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "トークンの生成中にエラーが発生しました")
		return
	}

	// レスポンスを返す
	c.JSON(http.StatusCreated, gin.H{
		"id":           user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"display_name": user.Name,
		"created_at":   user.CreatedAt,
		"token":        token,
	})
}

// LoginRequest ログインリクエストの構造体
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login ログインハンドラー
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// メールアドレスでユーザーを検索
	user, err := h.userRepo.GetByEmail(c, req.Email)
	if err != nil {
		h.log.Error("ユーザーの取得中にエラーが発生しました", "error", err)
		response.Unauthorized(c, "メールアドレスまたはパスワードが正しくありません")
		return
	}

	// パスワードを検証
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.log.Info("パスワードの検証に失敗しました", "userID", user.ID)
		response.Unauthorized(c, "メールアドレスまたはパスワードが正しくありません")
		return
	}

	// JWTトークンを生成
	token, err := h.jwtUtil.GenerateToken(user.ID.String())
	if err != nil {
		h.log.Error("トークンの生成中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "トークンの生成中にエラーが発生しました")
		return
	}

	// レスポンスを返す
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"display_name": user.Name,
			"avatar_url":   user.ProfileImage,
			"bio":          user.Bio,
		},
		"token": token,
	})
}

// RefreshToken トークン更新ハンドラー
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// リフレッシュトークンはAuthミドルウェアで検証済み
	// c.GetFromContextで現在のユーザーIDを取得
	userIDStr, exists := c.Get("userID")
	if !exists {
		h.log.Error("ユーザーIDがコンテキストに存在しません")
		response.Unauthorized(c, "トークンが無効です")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		h.log.Error("ユーザーIDの解析中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "トークンの更新中にエラーが発生しました")
		return
	}

	// ユーザーが存在するか確認
	_, err = h.userRepo.GetByID(c, userID)
	if err != nil {
		h.log.Error("ユーザーの確認中にエラーが発生しました", "error", err)
		response.Unauthorized(c, "トークンが無効です")
		return
	}

	// 新しいJWTトークンを生成
	token, err := h.jwtUtil.GenerateToken(userID.String())
	if err != nil {
		h.log.Error("トークンの生成中にエラーが発生しました", "error", err)
		response.InternalServerError(c, "トークンの生成中にエラーが発生しました")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// Logout ログアウトハンドラー
func (h *AuthHandler) Logout(c *gin.Context) {
	// サーバー側でトークンを無効化する必要はありません
	// クライアント側でトークンを削除すればOK
	// 必要に応じてブラックリストなどの仕組みを実装することも可能

	c.Status(http.StatusNoContent)
}
