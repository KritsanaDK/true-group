package repository

import (
	"context"
	"tdg/internal/model"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ITrueRepository interface {
	GetUserByID(userID int64) (*model.User, error)
	GetWatchHistory(userID int64, limit int) ([]model.WatchHistory, error)
	GetCandidateContent(userID int64, limit int) ([]model.Content, error)
	GetUsersBatch(page, limit int) ([]model.User, int, error)
}

type trueRepository struct {
	ctx    context.Context
	client *pgxpool.Pool
}

func NewTrueRepository(ctx context.Context, client *pgxpool.Pool) (ITrueRepository, error) {
	return &trueRepository{
		ctx:    ctx,
		client: client,
	}, nil
}

func (r *trueRepository) GetUserByID(userID int64) (*model.User, error) {

	query := `
	SELECT id, age, country, subscription_type
	FROM users
	WHERE id = $1
	`

	row := r.client.QueryRow(r.ctx, query, userID)

	user := model.User{}

	err := row.Scan(
		&user.ID,
		&user.Age,
		&user.Country,
		&user.SubscriptionType,
	)

	if err != nil {
		// "no rows" is a valid state for this method; callers decide how to handle missing users.
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (r *trueRepository) GetWatchHistory(userID int64, limit int) ([]model.WatchHistory, error) {

	query := `
		SELECT c.id, c.genre, uwh.watched_at
		FROM user_watch_history uwh
		JOIN content c ON uwh.content_id = c.id
		WHERE uwh.user_id = $1
		ORDER BY uwh.watched_at DESC
		LIMIT $2;
	`

	rows, err := r.client.Query(r.ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []model.WatchHistory

	for rows.Next() {

		var h model.WatchHistory

		err := rows.Scan(
			&h.ID,
			&h.Genre,
			&h.WatchedAt,
		)

		if err != nil {
			return nil, err
		}

		history = append(history, h)
	}

	return history, nil
}

func (r *trueRepository) GetCandidateContent(userID int64, limit int) ([]model.Content, error) {

	query := `
	SELECT id, title, genre, popularity_score, created_at
	FROM content
	WHERE id NOT IN (
		SELECT content_id
		FROM user_watch_history
		WHERE user_id = $1
	)
	ORDER BY popularity_score DESC
	LIMIT $2
	`

	rows, err := r.client.Query(r.ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []model.Content

	for rows.Next() {

		var c model.Content

		err := rows.Scan(
			&c.ID,
			&c.Title,
			&c.Genre,
			&c.PopularityScore,
			&c.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		contents = append(contents, c)
	}

	return contents, nil
}

func (r *trueRepository) GetUsersBatch(page, limit int) ([]model.User, int, error) {

	offset := (page - 1) * limit

	rows, err := r.client.Query(r.ctx,
		`SELECT id, age, country, subscription_type
		 FROM users
		 ORDER BY id
		 LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []model.User

	for rows.Next() {
		var u model.User
		err := rows.Scan(
			&u.ID,
			&u.Age,
			&u.Country,
			&u.SubscriptionType,
		)
		if err != nil {
			return nil, 0, err
		}

		users = append(users, u)
	}

	var total int
	err = r.client.QueryRow(r.ctx,
		`SELECT COUNT(*) FROM users`).Scan(&total)

	return users, total, err
}
