package postgres

import (
	"context"
	"errors"

	"github.com/TakuyaAizawa/gox/internal/repository/interfaces"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type followRepository struct {
	db *pgxpool.Pool
}

// NewFollowRepository creates a new PostgreSQL implementation of FollowRepository
func NewFollowRepository(db *pgxpool.Pool) interfaces.FollowRepository {
	return &followRepository{db: db}
}

func (r *followRepository) Follow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	// 自分自身をフォローできないようにする
	if followerID == followeeID {
		return errors.New("cannot follow yourself")
	}

	query := `
		INSERT INTO follows (follower_id, followee_id, created_at)
		VALUES ($1, $2, NOW())
	`

	_, err := r.db.Exec(ctx, query, followerID, followeeID)
	if err != nil {
		return err
	}

	// フォロワー数とフォロー数を更新
	updateFollowerCount := `
		UPDATE users SET follower_count = follower_count + 1
		WHERE id = $1
	`
	updateFollowingCount := `
		UPDATE users SET following_count = following_count + 1
		WHERE id = $1
	`

	_, err = r.db.Exec(ctx, updateFollowerCount, followeeID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, updateFollowingCount, followerID)
	if err != nil {
		return err
	}

	return nil
}

func (r *followRepository) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	query := `
		DELETE FROM follows
		WHERE follower_id = $1 AND followee_id = $2
	`

	result, err := r.db.Exec(ctx, query, followerID, followeeID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("follow relationship not found")
	}

	// フォロワー数とフォロー数を更新
	updateFollowerCount := `
		UPDATE users SET follower_count = GREATEST(follower_count - 1, 0)
		WHERE id = $1
	`
	updateFollowingCount := `
		UPDATE users SET following_count = GREATEST(following_count - 1, 0)
		WHERE id = $1
	`

	_, err = r.db.Exec(ctx, updateFollowerCount, followeeID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, updateFollowingCount, followerID)
	if err != nil {
		return err
	}

	return nil
}

func (r *followRepository) IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM follows
			WHERE follower_id = $1 AND followee_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, followerID, followeeID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *followRepository) GetFollowers(ctx context.Context, userID uuid.UUID, offset, limit int) ([]uuid.UUID, error) {
	query := `
		SELECT follower_id FROM follows
		WHERE followee_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followers []uuid.UUID
	for rows.Next() {
		var followerID uuid.UUID
		if err := rows.Scan(&followerID); err != nil {
			return nil, err
		}
		followers = append(followers, followerID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return followers, nil
}

func (r *followRepository) GetFollowing(ctx context.Context, userID uuid.UUID, offset, limit int) ([]uuid.UUID, error) {
	query := `
		SELECT followee_id FROM follows
		WHERE follower_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var following []uuid.UUID
	for rows.Next() {
		var followeeID uuid.UUID
		if err := rows.Scan(&followeeID); err != nil {
			return nil, err
		}
		following = append(following, followeeID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return following, nil
}

func (r *followRepository) CountFollowers(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := "SELECT COUNT(*) FROM follows WHERE followee_id = $1"

	var count int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *followRepository) CountFollowing(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := "SELECT COUNT(*) FROM follows WHERE follower_id = $1"

	var count int64
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
