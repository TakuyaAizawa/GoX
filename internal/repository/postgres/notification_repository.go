package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/TakuyaAizawa/gox/internal/domain/models"
	"github.com/TakuyaAizawa/gox/internal/repository/interfaces"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type notificationRepository struct {
	db *pgxpool.Pool
}

// NewNotificationRepository creates a new PostgreSQL implementation of NotificationRepository
func NewNotificationRepository(db *pgxpool.Pool) interfaces.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	query := `
		INSERT INTO notifications (
			id, user_id, actor_id, type, post_id, is_read, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query,
		notification.ID, notification.UserID, notification.ActorID,
		notification.Type, notification.PostID, notification.IsRead,
		notification.CreatedAt,
	)

	return err
}

func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	query := `
		SELECT id, user_id, actor_id, type, post_id, is_read, created_at
		FROM notifications WHERE id = $1
	`

	notification := &models.Notification{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&notification.ID, &notification.UserID, &notification.ActorID,
		&notification.Type, &notification.PostID, &notification.IsRead,
		&notification.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return notification, nil
}

func (r *notificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.Notification, error) {
	query := `
		SELECT id, user_id, actor_id, type, post_id, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		notification := &models.Notification{}
		err := rows.Scan(
			&notification.ID, &notification.UserID, &notification.ActorID,
			&notification.Type, &notification.PostID, &notification.IsRead,
			&notification.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = true
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("notification not found")
	}

	return nil
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = true
		WHERE user_id = $1 AND is_read = false
	`

	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *notificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM notifications WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("notification not found")
	}

	return nil
}

func (r *notificationRepository) CountUnreadByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := "SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false"

	var count int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *notificationRepository) GetWithRelations(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	query := `
		WITH notification_data AS (
			SELECT n.*, 
				u.username as actor_username, u.email as actor_email,
				u.name as actor_name, u.bio as actor_bio,
				u.profile_image as actor_profile_image,
				u.follower_count as actor_follower_count,
				u.following_count as actor_following_count,
				u.post_count as actor_post_count,
				u.is_verified as actor_is_verified,
				u.created_at as actor_created_at,
				p.user_id as post_user_id, p.content as post_content,
				p.media_urls as post_media_urls,
				p.like_count as post_like_count,
				p.repost_count as post_repost_count,
				p.reply_count as post_reply_count,
				p.is_repost as post_is_repost,
				p.repost_id as post_repost_id,
				p.is_reply as post_is_reply,
				p.reply_to_id as post_reply_to_id,
				p.created_at as post_created_at,
				p.updated_at as post_updated_at
			FROM notifications n
			LEFT JOIN users u ON n.actor_id = u.id
			LEFT JOIN posts p ON n.post_id = p.id
			WHERE n.id = $1
		)
		SELECT * FROM notification_data
	`

	notification := &models.Notification{}
	actor := &models.User{}
	post := &models.Post{}

	var (
		actorUsername, actorEmail, actorName           *string
		actorBio, actorProfileImage                    *string
		actorFollowerCount, actorFollowingCount        *int
		actorPostCount                                 *int
		actorIsVerified                                *bool
		actorCreatedAt, postCreatedAt, postUpdatedAt   *time.Time
		postUserID, postRepostID, postReplyToID        *uuid.UUID
		postContent                                    *string
		postMediaURLsJSON                              []byte
		postLikeCount, postRepostCount, postReplyCount *int
		postIsRepost, postIsReply                      *bool
	)

	err := r.db.QueryRow(ctx, query, id).Scan(
		&notification.ID, &notification.UserID, &notification.ActorID,
		&notification.Type, &notification.PostID, &notification.IsRead,
		&notification.CreatedAt,
		&actorUsername, &actorEmail, &actorName, &actorBio,
		&actorProfileImage, &actorFollowerCount, &actorFollowingCount,
		&actorPostCount, &actorIsVerified, &actorCreatedAt,
		&postUserID, &postContent, &postMediaURLsJSON,
		&postLikeCount, &postRepostCount, &postReplyCount,
		&postIsRepost, &postRepostID, &postIsReply,
		&postReplyToID, &postCreatedAt, &postUpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if actorUsername != nil {
		actor.ID = notification.ActorID
		actor.Username = *actorUsername
		actor.Email = *actorEmail
		actor.Name = *actorName
		actor.Bio = *actorBio
		actor.ProfileImage = *actorProfileImage
		actor.FollowerCount = *actorFollowerCount
		actor.FollowingCount = *actorFollowingCount
		actor.PostCount = *actorPostCount
		actor.IsVerified = *actorIsVerified
		actor.CreatedAt = *actorCreatedAt
		actor.UpdatedAt = *actorCreatedAt
		notification.Actor = actor.ToResponse()
	}

	if notification.PostID != nil && postContent != nil {
		post.ID = *notification.PostID
		post.UserID = *postUserID
		post.Content = *postContent
		if postMediaURLsJSON != nil {
			if err := json.Unmarshal(postMediaURLsJSON, &post.MediaURLs); err != nil {
				return nil, err
			}
		}
		post.LikeCount = *postLikeCount
		post.RepostCount = *postRepostCount
		post.ReplyCount = *postReplyCount
		post.IsRepost = *postIsRepost
		post.RepostID = postRepostID
		post.IsReply = *postIsReply
		post.ReplyToID = postReplyToID
		post.CreatedAt = *postCreatedAt
		post.UpdatedAt = *postUpdatedAt
		notification.Post = post.ToResponse()
	}

	return notification, nil
}

func (r *notificationRepository) GetByUserIDWithRelations(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.Notification, error) {
	query := `
		WITH notification_data AS (
			SELECT n.*, 
				u.username as actor_username, u.email as actor_email,
				u.name as actor_name, u.bio as actor_bio,
				u.profile_image as actor_profile_image,
				u.follower_count as actor_follower_count,
				u.following_count as actor_following_count,
				u.post_count as actor_post_count,
				u.is_verified as actor_is_verified,
				u.created_at as actor_created_at,
				p.user_id as post_user_id, p.content as post_content,
				p.media_urls as post_media_urls,
				p.like_count as post_like_count,
				p.repost_count as post_repost_count,
				p.reply_count as post_reply_count,
				p.is_repost as post_is_repost,
				p.repost_id as post_repost_id,
				p.is_reply as post_is_reply,
				p.reply_to_id as post_reply_to_id,
				p.created_at as post_created_at,
				p.updated_at as post_updated_at
			FROM notifications n
			LEFT JOIN users u ON n.actor_id = u.id
			LEFT JOIN posts p ON n.post_id = p.id
			WHERE n.user_id = $1
			ORDER BY n.created_at DESC
			LIMIT $2 OFFSET $3
		)
		SELECT * FROM notification_data
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		notification := &models.Notification{}
		actor := &models.User{}
		post := &models.Post{}

		var (
			actorUsername, actorEmail, actorName           *string
			actorBio, actorProfileImage                    *string
			actorFollowerCount, actorFollowingCount        *int
			actorPostCount                                 *int
			actorIsVerified                                *bool
			actorCreatedAt, postCreatedAt, postUpdatedAt   *time.Time
			postUserID, postRepostID, postReplyToID        *uuid.UUID
			postContent                                    *string
			postMediaURLsJSON                              []byte
			postLikeCount, postRepostCount, postReplyCount *int
			postIsRepost, postIsReply                      *bool
		)

		err := rows.Scan(
			&notification.ID, &notification.UserID, &notification.ActorID,
			&notification.Type, &notification.PostID, &notification.IsRead,
			&notification.CreatedAt,
			&actorUsername, &actorEmail, &actorName, &actorBio,
			&actorProfileImage, &actorFollowerCount, &actorFollowingCount,
			&actorPostCount, &actorIsVerified, &actorCreatedAt,
			&postUserID, &postContent, &postMediaURLsJSON,
			&postLikeCount, &postRepostCount, &postReplyCount,
			&postIsRepost, &postRepostID, &postIsReply,
			&postReplyToID, &postCreatedAt, &postUpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if actorUsername != nil {
			actor.ID = notification.ActorID
			actor.Username = *actorUsername
			actor.Email = *actorEmail
			actor.Name = *actorName
			actor.Bio = *actorBio
			actor.ProfileImage = *actorProfileImage
			actor.FollowerCount = *actorFollowerCount
			actor.FollowingCount = *actorFollowingCount
			actor.PostCount = *actorPostCount
			actor.IsVerified = *actorIsVerified
			actor.CreatedAt = *actorCreatedAt
			actor.UpdatedAt = *actorCreatedAt
			notification.Actor = actor.ToResponse()
		}

		if notification.PostID != nil && postContent != nil {
			post.ID = *notification.PostID
			post.UserID = *postUserID
			post.Content = *postContent
			if postMediaURLsJSON != nil {
				if err := json.Unmarshal(postMediaURLsJSON, &post.MediaURLs); err != nil {
					return nil, err
				}
			}
			post.LikeCount = *postLikeCount
			post.RepostCount = *postRepostCount
			post.ReplyCount = *postReplyCount
			post.IsRepost = *postIsRepost
			post.RepostID = postRepostID
			post.IsReply = *postIsReply
			post.ReplyToID = postReplyToID
			post.CreatedAt = *postCreatedAt
			post.UpdatedAt = *postUpdatedAt
			notification.Post = post.ToResponse()
		}

		notifications = append(notifications, notification)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}
