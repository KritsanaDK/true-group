package utils

import (
	"fmt"
	"math/rand/v2"
	"tdg/internal/model"
	"time"
)

func BytesToString(data []byte) string {
	if data == nil {
		return ""
	}
	return string(data[:])
}

func StringToBytes(s string) []byte {
	return []byte(s)
}

func ParseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
func CountGenres(history []model.WatchHistory) map[string]int {

	genreCounts := make(map[string]int)

	for _, item := range history {
		genreCounts[item.Genre]++
	}

	return genreCounts
}

func NormalizeGenreCounts(counts map[string]int) map[string]float64 {

	total := 0
	for _, count := range counts {
		total += count
	}

	prefs := make(map[string]float64)

	if total == 0 {
		return prefs
	}

	for genre, count := range counts {
		prefs[genre] = float64(count) / float64(total)
	}

	return prefs
}

func CalculateScore(
	content model.Content,
	genrePreferences map[string]float64,
) float64 {

	// popularity component
	popularityComponent := content.PopularityScore * 0.4

	// genre boost
	genrePreference := genrePreferences[content.Genre]
	if genrePreference == 0 {
		genrePreference = 0.1
	}

	genreBoost := genrePreference * 0.35

	// recency
	recencyFactor := CalculateRecencyFactor(content.CreatedAt)
	recencyComponent := recencyFactor * 0.15

	// exploration randomness
	randomNoise := (rand.Float64()*0.1 - 0.05) * 0.1

	finalScore :=
		popularityComponent +
			genreBoost +
			recencyComponent +
			randomNoise

	return finalScore
}

func CalculateRecencyFactor(createdAt time.Time) float64 {
	now := time.Now()

	days := now.Sub(createdAt).Hours() / 24

	recencyFactor := 1.0 / (1.0 + (days / 365.0))

	return recencyFactor
}
