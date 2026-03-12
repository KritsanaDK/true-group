package repository

import (
	"context"
	"tdg/internal/model"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const defaultQueryTimeout = 3 * time.Second

type ITrueRepository interface {
	GetUserByID(userID int64) (*model.User, error)
	GetWatchHistory(userID int64, limit int) ([]model.WatchHistory, error)
	GetWatchHistoryByUserIDs(userIDs []int64, limit int) (map[int64][]model.WatchHistory, error)
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

func (r *trueRepository) withTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.ctx, defaultQueryTimeout)
}

func (r *trueRepository) GetUserByID(userID int64) (*model.User, error) {

	query := `
	SELECT id, age, country, subscription_type
	FROM users
	WHERE id = $1
	`

	queryCtx, cancel := r.withTimeout()
	defer cancel()

	row := r.client.QueryRow(queryCtx, query, userID)

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

	queryCtx, cancel := r.withTimeout()
	defer cancel()

	rows, err := r.client.Query(queryCtx, query, userID, limit)
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

func (r *trueRepository) GetWatchHistoryByUserIDs(userIDs []int64, limit int) (map[int64][]model.WatchHistory, error) {
	historyByUser := make(map[int64][]model.WatchHistory)

	if len(userIDs) == 0 {
		return historyByUser, nil
	}

	query := `
		SELECT user_id, content_id, genre, watched_at
		FROM (
			SELECT
				uwh.user_id,
				c.id AS content_id,
				c.genre,
				uwh.watched_at,
				ROW_NUMBER() OVER (PARTITION BY uwh.user_id ORDER BY uwh.watched_at DESC) AS rn
			FROM user_watch_history uwh
			JOIN content c ON uwh.content_id = c.id
			WHERE uwh.user_id = ANY($1)
		) ranked
		WHERE rn <= $2
		ORDER BY user_id, watched_at DESC
	`

	queryCtx, cancel := r.withTimeout()
	defer cancel()

	rows, err := r.client.Query(queryCtx, query, userIDs, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID int64
		var item model.WatchHistory

		err := rows.Scan(
			&userID,
			&item.ID,
			&item.Genre,
			&item.WatchedAt,
		)
		if err != nil {
			return nil, err
		}

		historyByUser[userID] = append(historyByUser[userID], item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return historyByUser, nil
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

	queryCtx, cancel := r.withTimeout()
	defer cancel()

	rows, err := r.client.Query(queryCtx, query, userID, limit)
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

	queryCtx, cancel := r.withTimeout()
	defer cancel()

	rows, err := r.client.Query(queryCtx,
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
	err = r.client.QueryRow(queryCtx,
		`SELECT COUNT(*) FROM users`).Scan(&total)

	return users, total, err
}
