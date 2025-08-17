package article

// Recommendation represents a recommended article with its average score.
type Recommendation struct {
	Article Article // Article struct exists in domain/article
	Score   float64 // Average rating score
}
