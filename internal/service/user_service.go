package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"tdg/internal/cache"
	"tdg/internal/model"
	"tdg/internal/repository"
	"tdg/internal/utils"
)

var ErrUserNotFound = errors.New("user not found")
var ErrModelUnavailable = errors.New("model unavailable")

type ITrueService interface {
	GetRecommendations(userID int64, limit int) (model.RecommendationResponse, error)
	GetBatchRecommendations(page, limit int) (model.BatchRecommendationResponse, error)
}

type trueService struct {
	ctx      context.Context
	debug    bool
	userRepo repository.ITrueRepository
	cache    cache.ICacheClient
}

func NewTrueService(
	ctx context.Context,
	debug bool,
	repo AllRepository,
) ITrueService {

	rand.Seed(time.Now().UnixNano())

	return &trueService{
		ctx:      ctx,
		debug:    debug,
		userRepo: repo.ITrueRepository,
		cache:    repo.ICacheClient,
	}
}

func (s *trueService) GetRecommendations(
	userID int64,
	limit int,
) (model.RecommendationResponse, error) {

	cacheKey := fmt.Sprintf("rec:user:%d:limit:%d", userID, limit)

	// --------------------------------------------------
	// 1. Cache Check
	// --------------------------------------------------

	cacheVal, err := s.cache.Get(cacheKey)

	if err == nil && cacheVal != "" {

		var cached model.RecommendationResponse

		if err := json.Unmarshal([]byte(cacheVal), &cached); err == nil {

			cached.Metadata.CacheHit = true
			return cached, nil
		}
	}

	// --------------------------------------------------
	// 2. Fetch User
	// --------------------------------------------------

	user, err := s.userRepo.GetUserByID(userID)

	if err != nil {
		return model.RecommendationResponse{}, err
	}

	if user == nil {
		return model.RecommendationResponse{}, ErrUserNotFound
	}

	// --------------------------------------------------
	// 3. Fetch Watch History
	// --------------------------------------------------

	history, err := s.userRepo.GetWatchHistory(userID, 50)

	if err != nil {
		return model.RecommendationResponse{}, err
	}

	// --------------------------------------------------
	// 4. Build Genre Preferences
	// --------------------------------------------------

	genreCounts := utils.CountGenres(history)

	genrePreferences := utils.NormalizeGenreCounts(genreCounts)

	// --------------------------------------------------
	// 5. Fetch Candidate Content
	// --------------------------------------------------

	candidates, err := s.userRepo.GetCandidateContent(user.ID, 100)

	if err != nil {
		return model.RecommendationResponse{}, err
	}

	// --------------------------------------------------
	// 6. Simulate Model Latency (30-50ms)
	// --------------------------------------------------

	delay := rand.Intn(21) + 30
	time.Sleep(time.Duration(delay) * time.Millisecond)

	// --------------------------------------------------
	// 7. Simulate Random Failure (1.5%)
	// --------------------------------------------------

	if rand.Float64() < 0.015 {
		return model.RecommendationResponse{}, ErrModelUnavailable
	}

	// --------------------------------------------------
	// 8. Score Candidates
	// --------------------------------------------------

	var scoredCandidates []model.Recommendation

	for _, content := range candidates {

		score := utils.CalculateScore(content, genrePreferences)

		rec := model.Recommendation{
			ContentID:       content.ID,
			Title:           content.Title,
			Genre:           content.Genre,
			PopularityScore: content.PopularityScore,
			Score:           score,
		}

		scoredCandidates = append(scoredCandidates, rec)
	}

	// --------------------------------------------------
	// 9. Sort by Score
	// --------------------------------------------------

	sort.Slice(scoredCandidates, func(i, j int) bool {
		return scoredCandidates[i].Score > scoredCandidates[j].Score
	})

	// --------------------------------------------------
	// 10. Apply Limit
	// --------------------------------------------------

	if len(scoredCandidates) > limit {
		scoredCandidates = scoredCandidates[:limit]
	}

	// --------------------------------------------------
	// 11. Build Response
	// --------------------------------------------------

	response := model.RecommendationResponse{
		UserID:          user.ID,
		Recommendations: scoredCandidates,
		Metadata: model.Metadata{
			CacheHit:    false,
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
			TotalCount:  len(scoredCandidates),
		},
	}

	// --------------------------------------------------
	// 12. Store Cache (TTL 10 minutes)
	// --------------------------------------------------

	bytes, _ := json.Marshal(response)

	_ = s.cache.Set(
		cacheKey,
		string(bytes),
		10*time.Minute,
	)

	return response, nil
}

func (s *trueService) GetBatchRecommendations(page, limit int) (model.BatchRecommendationResponse, error) {

	start := time.Now()

	users, totalUsers, err := s.userRepo.GetUsersBatch(page, limit)
	if err != nil {
		return model.BatchRecommendationResponse{}, err
	}

	workerCount := 5

	jobs := make(chan model.User, len(users))
	results := make(chan model.BatchRecommendationResult, len(users))

	var wg sync.WaitGroup

	// Worker
	worker := func() {
		defer wg.Done()

		for user := range jobs {

			recs, err := s.GetRecommendations(user.ID, 100)

			if err != nil {

				results <- model.BatchRecommendationResult{
					UserID:  user.ID,
					Status:  "failed",
					Error:   "model_inference_timeout",
					Message: err.Error(),
				}

				continue
			}

			results <- model.BatchRecommendationResult{
				UserID:          user.ID,
				Recommendations: recs.Recommendations,
				Status:          "success",
			}
		}
	}

	// start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker()
	}

	// send jobs
	for _, u := range users {
		jobs <- u
	}

	close(jobs)

	wg.Wait()
	close(results)

	var responseResults []model.BatchRecommendationResult
	success := 0
	failed := 0

	for r := range results {

		responseResults = append(responseResults, r)

		if r.Status == "success" {
			success++
		} else {
			failed++
		}
	}

	processingTime := time.Since(start).Milliseconds()

	return model.BatchRecommendationResponse{
		Page:       page,
		Limit:      limit,
		TotalUsers: totalUsers,
		Results:    responseResults,
		Summary: model.BatchSummary{
			SuccessCount:     success,
			FailedCount:      failed,
			ProcessingTimeMS: processingTime,
		},
		Metadata: model.BatchMetadata{
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}, nil
}
