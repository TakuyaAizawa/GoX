package postgres

import (
	"context"
	"errors"

	"github.com/TakuyaAizawa/gox/internal/domain/models"
	"github.com/TakuyaAizawa/gox/internal/repository/interfaces"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type likeRepository struct {
	db *pgxpool.Pool
}

// NewLikeRepository creates a new PostgreSQL implementation of LikeRepository
func NewLikeRepository(db *pgxpool.Pool) interfaces.LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) Like(ctx context.Context, like *models.Like) error {
	query := `
		INSERT INTO likes (user_id, post_id, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(ctx, query, like.UserID, like.PostID, like.CreatedAt)
	if err != nil {
		return err
	}

	// いいね数を更新
	updateLikeCount := `
		UPDATE posts SET like_count = like_count + 1
		WHERE id = $1
	`

	_, err = r.db.Exec(ctx, updateLikeCount, like.PostID)
	if err != nil {
		return err
	}

	return nil
}

func (r *likeRepository) Unlike(ctx context.Context, userID, postID uuid.UUID) error {
	query := `
		DELETE FROM likes
		WHERE user_id = $1 AND post_id = $2
	`

	result, err := r.db.Exec(ctx, query, userID, postID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("like relationship not found")
	}

	// いいね数を更新
	updateLikeCount := `
		UPDATE posts SET like_count = GREATEST(like_count - 1, 0)
		WHERE id = $1
	`

	_, err = r.db.Exec(ctx, updateLikeCount, postID)
	if err != nil {
		return err
	}

	return nil
}

func (r *likeRepository) HasLiked(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM likes
			WHERE user_id = $1 AND post_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, userID, postID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *likeRepository) GetLikesByPostID(ctx context.Context, postID uuid.UUID, offset, limit int) ([]*models.Like, error) {
	query := `
		SELECT user_id, post_id, created_at
		FROM likes
		WHERE post_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var likes []*models.Like
	for rows.Next() {
		like := &models.Like{}
		if err := rows.Scan(&like.UserID, &like.PostID, &like.CreatedAt); err != nil {
			return nil, err
		}
		likes = append(likes, like)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return likes, nil
}

func (r *likeRepository) GetLikesByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.Like, error) {
	query := `
		SELECT user_id, post_id, created_at
		FROM likes
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var likes []*models.Like
	for rows.Next() {
		like := &models.Like{}
		if err := rows.Scan(&like.UserID, &like.PostID, &like.CreatedAt); err != nil {
			return nil, err
		}
		likes = append(likes, like)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return likes, nil
}

func (r *likeRepository) CountLikesByPostID(ctx context.Context, postID uuid.UUID) (int64, error) {
	query := "SELECT COUNT(*) FROM likes WHERE post_id = $1"

	var count int64
	err := r.db.QueryRow(ctx, query, postID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *likeRepository) CountLikesByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := "SELECT COUNT(*) FROM likes WHERE user_id = $1"

	var count int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}