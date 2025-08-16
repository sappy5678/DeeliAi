package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/chatbotgang/go-clean-architecture-template/internal/domain/article"
	"github.com/chatbotgang/go-clean-architecture-template/testdata"
)

func assertArticle(t *testing.T, expected *article.Article, actual *article.Article) {
	require.NotNil(t, actual)
	assert.Equal(t, expected.URL, actual.URL)
	assert.Equal(t, expected.Title, actual.Title)
	assert.Equal(t, expected.Description, actual.Description)
	assert.Equal(t, expected.ImageURL, actual.ImageURL)
}

func TestPostgresRepository_CreateArticle(t *testing.T) {
	db := getTestPostgresDB()
	repo := initRepository(t, db, testdata.Path(testdata.TestDataUser))

	url := "https://example.com/new-article"

	art, err := repo.CreateArticle(context.Background(), url)
	require.NoError(t, err)

	assert.Equal(t, url, art.URL)
	assert.NotEqual(t, uuid.Nil, art.ID)
}

func TestPostgresRepository_CreateUserArticle(t *testing.T) {
	db := getTestPostgresDB()
	repo := initRepository(t, db, testdata.Path(testdata.TestDataUser), testdata.Path(testdata.TestDataArticle))

	userID := uuid.MustParse("8a6d6322-a428-441a-8433-9551b1048a8c")
	articleID := uuid.MustParse("1a7a3f16-6e3b-4758-9882-03d4a7d0251b")

	userArticle, err := repo.CreateUserArticle(context.Background(), userID, articleID)
	require.NoError(t, err)

	assert.Equal(t, userID, userArticle.UserID)
	assert.Equal(t, articleID, userArticle.ArticleID)
}

func TestPostgresRepository_GetUserArticle(t *testing.T) {
	db := getTestPostgresDB()
	repo := initRepository(t, db, testdata.Path(testdata.TestDataUser), testdata.Path(testdata.TestDataArticle), testdata.Path(testdata.TestDataUserArticle))

	userID := uuid.MustParse("8a6d6322-a428-441a-8433-9551b1048a8c")
	articleID := uuid.MustParse("1a7a3f16-6e3b-4758-9882-03d4a7d0251b")

	userArticle, err := repo.GetUserArticle(context.Background(), userID, articleID)
	require.NoError(t, err)

	assert.Equal(t, userID, userArticle.UserID)
	assert.Equal(t, articleID, userArticle.ArticleID)
	assert.Equal(t, int16(5), userArticle.Rate)
}

func TestPostgresRepository_ListArticles(t *testing.T) {
	db := getTestPostgresDB()
	repo := initRepository(t, db, testdata.Path(testdata.TestDataUser), testdata.Path(testdata.TestDataArticle), testdata.Path(testdata.TestDataUserArticle))

	userID := uuid.MustParse("8a6d6322-a428-441a-8433-9551b1048a8c")

	articles, err := repo.ListArticles(context.Background(), userID, uuid.Nil, 10)
	require.NoError(t, err)
	assert.Len(t, articles, 2)
}

func TestPostgresRepository_DeleteUserArticle(t *testing.T) {
	db := getTestPostgresDB()
	repo := initRepository(t, db, testdata.Path(testdata.TestDataUser), testdata.Path(testdata.TestDataArticle), testdata.Path(testdata.TestDataUserArticle))

	userID := uuid.MustParse("8a6d6322-a428-441a-8433-9551b1048a8c")
	articleID := uuid.MustParse("1a7a3f16-6e3b-4758-9882-03d4a7d0251b")

	err := repo.DeleteUserArticle(context.Background(), userID, articleID)
	require.NoError(t, err)

	_, err = repo.GetUserArticle(context.Background(), userID, articleID)
	require.Error(t, err)
}

func TestPostgresRepository_UpdateUserArticleRate(t *testing.T) {
	db := getTestPostgresDB()
	repo := initRepository(t, db, testdata.Path(testdata.TestDataUser), testdata.Path(testdata.TestDataArticle), testdata.Path(testdata.TestDataUserArticle))

	userID := uuid.MustParse("8a6d6322-a428-441a-8433-9551b1048a8c")
	articleID := uuid.MustParse("1a7a3f16-6e3b-4758-9882-03d4a7d0251b")
	newRate := int16(3)

	err := repo.UpdateUserArticleRate(context.Background(), userID, articleID, newRate)
	require.NoError(t, err)

	userArticle, err := repo.GetUserArticle(context.Background(), userID, articleID)
	require.NoError(t, err)
	assert.Equal(t, newRate, userArticle.Rate)
}

func TestPostgresRepository_DeleteUserArticleRate(t *testing.T) {
	db := getTestPostgresDB()
	repo := initRepository(t, db, testdata.Path(testdata.TestDataUser), testdata.Path(testdata.TestDataArticle), testdata.Path(testdata.TestDataUserArticle))

	userID := uuid.MustParse("8a6d6322-a428-441a-8433-9551b1048a8c")
	articleID := uuid.MustParse("1a7a3f16-6e3b-4758-9882-03d4a7d0251b")

	err := repo.DeleteUserArticleRate(context.Background(), userID, articleID)
	require.NoError(t, err)

	userArticle, err := repo.GetUserArticle(context.Background(), userID, articleID)
	require.NoError(t, err)
	assert.Equal(t, int16(0), userArticle.Rate)
}
