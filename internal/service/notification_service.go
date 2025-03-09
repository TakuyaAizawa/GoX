package service

import (
	"context"
	"fmt"

	"github.com/TakuyaAizawa/gox/internal/domain/models"
	"github.com/TakuyaAizawa/gox/internal/repository/interfaces"
	"github.com/TakuyaAizawa/gox/internal/websocket"
	"github.com/TakuyaAizawa/gox/pkg/logger"
	"github.com/google/uuid"
)

// NotificationService 通知関連のビジネスロジックを管理するサービス
type NotificationService struct {
	notificationRepo interfaces.NotificationRepository
	userRepo         interfaces.UserRepository
	postRepo         interfaces.PostRepository
	hub              *websocket.Hub
	log              logger.Logger
}

// NewNotificationService 新しい通知サービスを作成する
func NewNotificationService(
	notificationRepo interfaces.NotificationRepository,
	userRepo interfaces.UserRepository,
	postRepo interfaces.PostRepository,
	hub *websocket.Hub,
	log logger.Logger,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		postRepo:         postRepo,
		hub:              hub,
		log:              log,
	}
}

// CreateLikeNotification いいね通知を作成する
func (s *NotificationService) CreateLikeNotification(ctx context.Context, actorID, recipientID uuid.UUID, postID uuid.UUID) error {
	// 自分自身へのいいねは通知しない
	if actorID == recipientID {
		return nil
	}

	// アクターユーザー情報の取得
	actor, err := s.userRepo.GetByID(ctx, actorID)
	if err != nil {
		s.log.Error("いいね通知: アクターユーザー取得エラー", "error", err)
		return err
	}

	// 投稿情報の取得
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		s.log.Error("いいね通知: 投稿取得エラー", "error", err)
		return err
	}

	// 通知レコードの作成
	notification := models.NewNotification(
		recipientID,
		actorID,
		models.NotificationTypeLike,
		&postID,
	)

	err = s.notificationRepo.Create(ctx, notification)
	if err != nil {
		s.log.Error("いいね通知: 保存エラー", "error", err)
		return err
	}

	// WebSocket通知の作成
	notificationEvent := websocket.NotificationEvent{
		ID:        notification.ID,
		Type:      websocket.EventTypeLike,
		CreatedAt: notification.CreatedAt,
		Message:   fmt.Sprintf("%sさんがあなたの投稿にいいねしました", actor.Name),
		Actor: websocket.ActorInfo{
			ID:          actor.ID,
			Username:    actor.Username,
			DisplayName: actor.Name,
			AvatarURL:   actor.ProfileImage,
		},
		Post: &websocket.PostInfo{
			ID:      post.ID,
			Content: truncateString(post.Content, 50),
		},
	}

	// WebSocketを通じて通知を送信
	message := websocket.NewNotificationMessage(notificationEvent)
	err = s.hub.NotifyUser(recipientID, message)
	if err != nil {
		s.log.Warn("WebSocket通知の送信に失敗しました", "error", err)
		// WebSocket送信の失敗は処理を続行
	}

	return nil
}

// CreateFollowNotification フォロー通知を作成する
func (s *NotificationService) CreateFollowNotification(ctx context.Context, actorID, recipientID uuid.UUID) error {
	// 自分自身へのフォローは通知しない
	if actorID == recipientID {
		return nil
	}

	// アクターユーザー情報の取得
	actor, err := s.userRepo.GetByID(ctx, actorID)
	if err != nil {
		s.log.Error("フォロー通知: アクターユーザー取得エラー", "error", err)
		return err
	}

	// 通知レコードの作成
	notification := models.NewNotification(
		recipientID,
		actorID,
		models.NotificationTypeFollow,
		nil,
	)

	err = s.notificationRepo.Create(ctx, notification)
	if err != nil {
		s.log.Error("フォロー通知: 保存エラー", "error", err)
		return err
	}

	// WebSocket通知の作成
	notificationEvent := websocket.NotificationEvent{
		ID:        notification.ID,
		Type:      websocket.EventTypeFollow,
		CreatedAt: notification.CreatedAt,
		Message:   fmt.Sprintf("%sさんがあなたをフォローしました", actor.Name),
		Actor: websocket.ActorInfo{
			ID:          actor.ID,
			Username:    actor.Username,
			DisplayName: actor.Name,
			AvatarURL:   actor.ProfileImage,
		},
	}

	// WebSocketを通じて通知を送信
	message := websocket.NewNotificationMessage(notificationEvent)
	err = s.hub.NotifyUser(recipientID, message)
	if err != nil {
		s.log.Warn("WebSocket通知の送信に失敗しました", "error", err)
		// WebSocket送信の失敗は処理を続行
	}

	return nil
}

// CreateReplyNotification 返信通知を作成する
func (s *NotificationService) CreateReplyNotification(ctx context.Context, actorID, recipientID uuid.UUID, postID, replyID uuid.UUID) error {
	// 自分自身への返信は通知しない
	if actorID == recipientID {
		return nil
	}

	// アクターユーザー情報の取得
	actor, err := s.userRepo.GetByID(ctx, actorID)
	if err != nil {
		s.log.Error("返信通知: アクターユーザー取得エラー", "error", err)
		return err
	}

	// 返信投稿の取得
	reply, err := s.postRepo.GetByID(ctx, replyID)
	if err != nil {
		s.log.Error("返信通知: 返信投稿取得エラー", "error", err)
		return err
	}

	// 通知レコードの作成
	notification := models.NewNotification(
		recipientID,
		actorID,
		models.NotificationTypeReply,
		&replyID,
	)

	err = s.notificationRepo.Create(ctx, notification)
	if err != nil {
		s.log.Error("返信通知: 保存エラー", "error", err)
		return err
	}

	// WebSocket通知の作成
	notificationEvent := websocket.NotificationEvent{
		ID:        notification.ID,
		Type:      websocket.EventTypeReply,
		CreatedAt: notification.CreatedAt,
		Message:   fmt.Sprintf("%sさんがあなたの投稿に返信しました", actor.Name),
		Actor: websocket.ActorInfo{
			ID:          actor.ID,
			Username:    actor.Username,
			DisplayName: actor.Name,
			AvatarURL:   actor.ProfileImage,
		},
		Post: &websocket.PostInfo{
			ID:      reply.ID,
			Content: truncateString(reply.Content, 50),
		},
	}

	// WebSocketを通じて通知を送信
	message := websocket.NewNotificationMessage(notificationEvent)
	err = s.hub.NotifyUser(recipientID, message)
	if err != nil {
		s.log.Warn("WebSocket通知の送信に失敗しました", "error", err)
		// WebSocket送信の失敗は処理を続行
	}

	return nil
}

// 文字列を指定の長さで切り詰める補助関数
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength-3] + "..."
}
