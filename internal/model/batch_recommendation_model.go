package model

type BatchRecommendationResult struct {
	UserID          int64            `json:"user_id"`
	Recommendations []Recommendation `json:"recommendations,omitempty"`
	Status          string           `json:"status"`
	Error           string           `json:"error,omitempty"`
	Message         string           `json:"message,omitempty"`
}

type BatchSummary struct {
	SuccessCount     int   `json:"success_count"`
	FailedCount      int   `json:"failed_count"`
	ProcessingTimeMS int64 `json:"processing_time_ms"`
}

type BatchMetadata struct {
	GeneratedAt string `json:"generated_at"`
}

type BatchRecommendationResponse struct {
	Page       int                         `json:"page"`
	Limit      int                         `json:"limit"`
	TotalUsers int                         `json:"total_users"`
	Results    []BatchRecommendationResult `json:"results"`
	Summary    BatchSummary                `json:"summary"`
	Metadata   BatchMetadata               `json:"metadata"`
}
