package model

import "time"

type User struct {
	ID               int64  `json:"id"`
	Age              int    `json:"age"`
	Country          string `json:"country"`
	SubscriptionType string `json:"subscription_type"`
}

type Content struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Genre           string    `json:"genre"`
	PopularityScore float64   `json:"popularity_score"`
	CreatedAt       time.Time `json:"created_at"`
}

type WatchHistory struct {
	ID        int64     `json:"id"`
	Genre     string    `json:"genre"`
	WatchedAt time.Time `json:"watched_at"`
}

type UserRecommendations struct {
	UserID          int64            `json:"user_id"`
	Recommendations []Recommendation `json:"recommendations"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type Recommendation struct {
	ContentID       int64   `json:"content_id"`
	Title           string  `json:"title"`
	Genre           string  `json:"genre"`
	PopularityScore float64 `json:"popularity_score"`
	Score           float64 `json:"score"`
}

type Metadata struct {
	CacheHit    bool   `json:"cache_hit"`
	GeneratedAt string `json:"generated_at"`
	TotalCount  int    `json:"total_count"`
}

type RecommendationResponse struct {
	UserID          int64            `json:"user_id"`
	Recommendations []Recommendation `json:"recommendations"`
	Metadata        Metadata         `json:"metadata"`
}
