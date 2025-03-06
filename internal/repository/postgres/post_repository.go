package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/TakuyaAizawa/gox/internal/domain/models"
	"github.com/TakuyaAizawa/gox/internal/repository/interfaces"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postRepository struct {
	db *pgxpool.Pool
}

// NewPostRepository creates a new PostgreSQL implementation of PostRepository
func NewPostRepository(db *pgxpool.Pool) interfaces.PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *models.Post) error {
	// バリデーションチェック
	if post == nil {
		return errors.New("post cannot be nil")
	}
	if post.Content == "" {
		return errors.New("content cannot be empty")
	}
	if len(post.Content) > 280 {
		return errors.New("content cannot exceed 280 characters")
	}
	if len(post.MediaURLs) > 4 {
		return errors.New("cannot have more than 4 media URLs")
	}

	query := `
		INSERT INTO posts (
			id, user_id, content, media_urls, reply_to_id, repost_id,
			like_count, repost_count, reply_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	mediaURLsJSON, err := json.Marshal(post.MediaURLs)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query,
		post.ID, post.UserID, post.Content, mediaURLsJSON,
		post.ReplyToID, post.RepostID, post.LikeCount,
		post.RepostCount, post.ReplyCount, post.CreatedAt, post.UpdatedAt,
	)

	return err
}

func (r *postRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	query := `
		SELECT id, user_id, content, media_urls, reply_to_id, repost_id,
			like_count, repost_count, reply_count, created_at, updated_at
		FROM posts WHERE id = $1
	`

	var post models.Post
	var mediaURLsJSON []byte
	err := r.db.QueryRow(ctx, query, id).Scan(
		&post.ID, &post.UserID, &post.Content, &mediaURLsJSON,
		&post.ReplyToID, &post.RepostID, &post.LikeCount,
		&post.RepostCount, &post.ReplyCount, &post.CreatedAt, &post.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("post not found")
	}
	if err != nil {
		return nil, err
	}

	if mediaURLsJSON != nil {
		err = json.Unmarshal(mediaURLsJSON, &post.MediaURLs)
		if err != nil {
			return nil, err
		}
	}

	post.IsReply = post.ReplyToID != nil
	post.IsRepost = post.RepostID != nil

	return &post, nil
}

func (r *postRepository) Update(ctx context.Context, post *models.Post) error {
	// バリデーションチェック
	if post == nil {
		return errors.New("post cannot be nil")
	}
	if post.Content == "" {
		return errors.New("content cannot be empty")
	}
	if len(post.Content) > 280 {
		return errors.New("content cannot exceed 280 characters")
	}
	if len(post.MediaURLs) > 4 {
		return errors.New("cannot have more than 4 media URLs")
	}

	query := `
		UPDATE posts SET
			content = $1, media_urls = $2, like_count = $3,
			repost_count = $4, reply_count = $5, updated_at = $6
		WHERE id = $7
	`

	mediaURLsJSON, err := json.Marshal(post.MediaURLs)
	if err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, query,
		post.Content, mediaURLsJSON, post.LikeCount,
		post.RepostCount, post.ReplyCount, post.UpdatedAt, post.ID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("post not found")
	}

	return nil
}

func (r *postRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := "DELETE FROM posts WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("post not found")
	}

	return nil
}

func (r *postRepository) List(ctx context.Context, offset, limit int) ([]*models.Post, error) {
	query := `
		SELECT id, user_id, content, media_urls, reply_to_id, repost_id,
			like_count, repost_count, reply_count, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	return r.queryPosts(ctx, query, limit, offset)
}

func (r *postRepository) GetByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.Post, error) {
	query := `
		SELECT id, user_id, content, media_urls, reply_to_id, repost_id,
			like_count, repost_count, reply_count, created_at, updated_at
		FROM posts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	return r.queryPosts(ctx, query, userID, limit, offset)
}

func (r *postRepository) GetReplies(ctx context.Context, postID uuid.UUID, offset, limit int) ([]*models.Post, error) {
	query := `
		SELECT id, user_id, content, media_urls, reply_to_id, repost_id,
			like_count, repost_count, reply_count, created_at, updated_at
		FROM posts
		WHERE reply_to_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	return r.queryPosts(ctx, query, postID, limit, offset)
}

func (r *postRepository) GetReposts(ctx context.Context, postID uuid.UUID, offset, limit int) ([]*models.Post, error) {
	query := `
		SELECT id, user_id, content, media_urls, reply_to_id, repost_id,
			like_count, repost_count, reply_count, created_at, updated_at
		FROM posts
		WHERE repost_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	return r.queryPosts(ctx, query, postID, limit, offset)
}

func (r *postRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := "SELECT COUNT(*) FROM posts WHERE user_id = $1"

	var count int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *postRepository) CountReplies(ctx context.Context, postID uuid.UUID) (int64, error) {
	query := "SELECT COUNT(*) FROM posts WHERE reply_to_id = $1"

	var count int64
	err := r.db.QueryRow(ctx, query, postID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *postRepository) CountReposts(ctx context.Context, postID uuid.UUID) (int64, error) {
	query := "SELECT COUNT(*) FROM posts WHERE repost_id = $1"

	var count int64
	err := r.db.QueryRow(ctx, query, postID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *postRepository) IncrementLikeCount(ctx context.Context, postID uuid.UUID) error {
	query := `
		UPDATE posts
		SET like_count = like_count + 1
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, postID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("post not found")
	}

	return nil
}

func (r *postRepository) DecrementLikeCount(ctx context.Context, postID uuid.UUID) error {
	query := `
		UPDATE posts
		SET like_count = GREATEST(like_count - 1, 0)
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, postID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("post not found")
	}

	return nil
}

func (r *postRepository) IncrementRepostCount(ctx context.Context, postID uuid.UUID) error {
	query := `
		UPDATE posts
		SET repost_count = repost_count + 1
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, postID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("post not found")
	}

	return nil
}

func (r *postRepository) DecrementRepostCount(ctx context.Context, postID uuid.UUID) error {
	query := `
		UPDATE posts
		SET repost_count = GREATEST(repost_count - 1, 0)
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, postID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("post not found")
	}

	return nil
}

func (r *postRepository) IncrementReplyCount(ctx context.Context, postID uuid.UUID) error {
	query := `
		UPDATE posts
		SET reply_count = reply_count + 1
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, postID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("post not found")
	}

	return nil
}

func (r *postRepository) DecrementReplyCount(ctx context.Context, postID uuid.UUID) error {
	query := `
		UPDATE posts
		SET reply_count = GREATEST(reply_count - 1, 0)
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, postID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("post not found")
	}

	return nil
}

// queryPosts is a helper function to execute queries that return post lists
func (r *postRepository) queryPosts(ctx context.Context, query string, args ...interface{}) ([]*models.Post, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		var mediaURLsJSON []byte
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Content, &mediaURLsJSON,
			&post.ReplyToID, &post.RepostID, &post.LikeCount,
			&post.RepostCount, &post.ReplyCount, &post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if mediaURLsJSON != nil {
			err = json.Unmarshal(mediaURLsJSON, &post.MediaURLs)
			if err != nil {
				return nil, err
			}
		}

		post.IsReply = post.ReplyToID != nil
		post.IsRepost = post.RepostID != nil

		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
