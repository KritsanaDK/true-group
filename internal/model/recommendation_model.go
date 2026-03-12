package model

import "time"

type ContentScore struct {
	ContentID int64   `json:"content_id"`
	Score     float64 `json:"score"`
}

type GenrePreference struct {
	Genre string
	Score float64
}

type CandidateContent struct {
	ID              int64
	Genre           string
	PopularityScore float64
	CreatedAt       time.Time
}

type WatchHistoryItem struct {
	ContentID int64
	Genre     string
	WatchedAt time.Time
}

type WatchHistoryWithContent struct {
	UserID    int64     `db:"user_id"`
	ContentID int64     `db:"content_id"`
	Genre     string    `db:"genre"`
	WatchedAt time.Time `db:"watched_at"`
}
