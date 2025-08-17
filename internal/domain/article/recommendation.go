package article

type RecommendationArticles []RecommendationArticle

type RecommendationArticle struct {
	Article       *Article
	AverageRating float64
}
