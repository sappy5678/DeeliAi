package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/sappy5678/DeeliAi/internal/domain/article"
	"github.com/sappy5678/DeeliAi/internal/domain/common"
)

// --- article table ---

type repoArticle struct {
	ID          uuid.UUID       `db:"id"`
	URL         string          `db:"url"`
	Title       sql.NullString  `db:"title"`
	Description sql.NullString  `db:"description"`
	ImageURL    sql.NullString  `db:"image_url"`
	Metadata    json.RawMessage `db:"metadata"`
}

func (a *repoArticle) toDomain() *article.Article {
	return &article.Article{
		ID:          a.ID,
		URL:         a.URL,
		Title:       a.Title.String,
		Description: a.Description.String,
		ImageURL:    a.ImageURL.String,
		Metadata:    a.Metadata,
	}
}

const (
	repoTableArticle                       = "articles"
	repoMaterializedViewArticleAverageRate = "materialized_articles_average_rate"
)

type repoColumnPatternArticle struct {
	ID            string
	URL           string
	Title         string
	Description   string
	ImageURL      string
	Metadata      string
	AverageRating string
}

var repoColumnArticle = repoColumnPatternArticle{
	ID:            "id",
	URL:           "url",
	Title:         "title",
	Description:   "description",
	ImageURL:      "image_url",
	Metadata:      "metadata",
	AverageRating: "average_rating",
}

func (c repoColumnPatternArticle) columns() string {
	col := []string{
		c.ID,
		c.URL,
		c.Title,
		c.Description,
		c.ImageURL,
		c.Metadata,
	}
	for i, v := range col {
		col[i] = fmt.Sprintf("%s.%s", repoTableArticle, v)
	}
	return strings.Join(col, ", ")
}

func (c repoColumnPatternArticle) materializedViewColumns() string {
	col := []string{
		c.ID,
		c.URL,
		c.Title,
		c.Description,
		c.ImageURL,
		c.Metadata,
	}
	for i, v := range col {
		col[i] = fmt.Sprintf("%s.%s", repoMaterializedViewArticleAverageRate, v)
	}
	return strings.Join(col, ", ")
}

// --- user_articles table ---

type repoUserArticle struct {
	ID          int64     `db:"id"`
	UserID      uuid.UUID `db:"user_id"`
	ArticleID   uuid.UUID `db:"article_id"`
	Rate        int16     `db:"rate"`
	CollectedAt time.Time `db:"collected_at"`
}

func (ua *repoUserArticle) toDomain() *article.UserArticle {
	return &article.UserArticle{
		ID:          ua.ID,
		UserID:      ua.UserID,
		ArticleID:   ua.ArticleID,
		Rate:        ua.Rate,
		CollectedAt: ua.CollectedAt,
	}
}

const repoTableUserArticle = "user_articles"

type repoColumnPatternUserArticle struct {
	ID          string
	UserID      string
	ArticleID   string
	Rate        string
	CollectedAt string
}

var repoColumnUserArticle = repoColumnPatternUserArticle{
	ID:          "id",
	UserID:      "user_id",
	ArticleID:   "article_id",
	Rate:        "rate",
	CollectedAt: "collected_at",
}

func (c repoColumnPatternUserArticle) columns() string {
	return strings.Join([]string{
		c.ID,
		c.UserID,
		c.ArticleID,
		c.Rate,
		c.CollectedAt,
	}, ", ")
}

// --- repository methods ---

func (r *PostgresRepository) CreateArticle(ctx context.Context, url string) (*article.Article, common.Error) {
	insertQuery, insertArgs, err := r.pgsq.Insert(repoTableArticle).
		Columns(repoColumnArticle.URL).
		Values(url).
		Suffix(fmt.Sprintf("ON CONFLICT (%s) DO NOTHING", repoColumnArticle.URL)).
		ToSql()
	if err != nil {
		return nil, common.NewError(common.ErrorCodeInternalProcess, errors.Wrap(err, "failed to build insert query for article"))
	}

	if _, err = r.db.ExecContext(ctx, insertQuery, insertArgs...); err != nil {
		return nil, common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to insert article"))
	}

	selectQuery, selectArgs, err := r.pgsq.Select(repoColumnArticle.columns()).
		From(repoTableArticle).
		Where(sq.Eq{repoColumnArticle.URL: url}).
		ToSql()
	if err != nil {
		return nil, common.NewError(common.ErrorCodeInternalProcess, errors.Wrap(err, "failed to build select query for article"))
	}

	var row repoArticle
	if err = r.db.GetContext(ctx, &row, selectQuery, selectArgs...); err != nil {
		r.logger(ctx).Error().Str("query", selectQuery).Err(err).Msg("failed to select article")
		return nil, common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to select article"))
	}

	return row.toDomain(), nil
}

func (r *PostgresRepository) CreateUserArticle(ctx context.Context, userID uuid.UUID, articleID uuid.UUID) (*article.UserArticle, common.Error) {
	insert := map[string]interface{}{
		repoColumnUserArticle.UserID:    userID,
		repoColumnUserArticle.ArticleID: articleID,
	}

	query, args, err := r.pgsq.Insert(repoTableUserArticle).
		SetMap(insert).
		Suffix(fmt.Sprintf("ON CONFLICT (%s, %s) DO NOTHING", repoColumnUserArticle.UserID, repoColumnUserArticle.ArticleID)).
		Suffix(fmt.Sprintf("RETURNING %s", repoColumnUserArticle.columns())).
		ToSql()
	if err != nil {
		return nil, common.NewError(common.ErrorCodeInternalProcess, errors.Wrap(err, "failed to build insert query for user_article"))
	}

	var row repoUserArticle
	if err = r.db.GetContext(ctx, &row, query, args...); err != nil {
		if err == sql.ErrNoRows {
			// If the row already exists, we get it here.
			return r.GetUserArticle(ctx, userID, articleID)
		}
		return nil, common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to insert user_article"))
	}

	return row.toDomain(), nil
}

func (r *PostgresRepository) GetUserArticle(ctx context.Context, userID uuid.UUID, articleID uuid.UUID) (*article.UserArticle, common.Error) {
	where := sq.And{
		sq.Eq{repoColumnUserArticle.UserID: userID},
		sq.Eq{repoColumnUserArticle.ArticleID: articleID},
	}

	query, args, err := r.pgsq.Select(repoColumnUserArticle.columns()).
		From(repoTableUserArticle).
		Where(where).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, common.NewError(common.ErrorCodeInternalProcess, errors.Wrap(err, "failed to build select query for user_article"))
	}

	var row repoUserArticle
	if err = r.db.GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewError(common.ErrorCodeResourceNotFound, err)
		}
		return nil, common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to select user_article"))
	}

	return row.toDomain(), nil
}

func (r *PostgresRepository) ListArticles(ctx context.Context, userID uuid.UUID, afterID uuid.UUID, limit int) ([]*article.Article, common.Error) {
	where := sq.And{
		sq.Eq{fmt.Sprintf("%s.%s", repoTableUserArticle, repoColumnUserArticle.UserID): userID},
	}
	if afterID != uuid.Nil {
		where = append(where, sq.Gt{fmt.Sprintf("%s.%s", repoTableArticle, repoColumnArticle.ID): afterID.String()})
	}

	query, args, err := r.pgsq.Select(repoColumnArticle.columns()).
		From(repoTableArticle).
		Join(fmt.Sprintf("%s ON %s.%s = %s.%s",
			repoTableUserArticle,
			repoTableUserArticle, repoColumnUserArticle.ArticleID,
			repoTableArticle, repoColumnArticle.ID)).
		Where(where).
		OrderBy(fmt.Sprintf("%s.%s ASC", repoTableArticle, repoColumnArticle.ID)).
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		r.logger(ctx).Error().Str("query", query).Err(err).Msg("failed to build select query for articles")
		return nil, common.NewError(common.ErrorCodeInternalProcess, errors.Wrap(err, "failed to build select query for articles"))
	}

	var rows []repoArticle
	if err = r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		r.logger(ctx).Error().Str("query", query).Err(err).Msg("failed to select articles")
		return nil, common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to select articles"))
	}

	articles := make([]*article.Article, 0, len(rows))
	for _, row := range rows {
		articles = append(articles, row.toDomain())
	}

	return articles, nil
}

func (r *PostgresRepository) DeleteUserArticle(ctx context.Context, userID uuid.UUID, articleID uuid.UUID) common.Error {
	where := sq.And{
		sq.Eq{repoColumnUserArticle.UserID: userID},
		sq.Eq{repoColumnUserArticle.ArticleID: articleID},
	}

	query, args, err := r.pgsq.Delete(repoTableUserArticle).
		Where(where).
		ToSql()
	if err != nil {
		return common.NewError(common.ErrorCodeInternalProcess, errors.Wrap(err, "failed to build delete query for user_article"))
	}

	if _, err = r.db.ExecContext(ctx, query, args...); err != nil {
		return common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to delete user_article"))
	}

	return nil
}

func (r *PostgresRepository) UpdateUserArticleRate(ctx context.Context, userID uuid.UUID, articleID uuid.UUID, rate int16) common.Error {
	update := map[string]interface{}{
		repoColumnUserArticle.Rate: rate,
	}

	where := sq.And{
		sq.Eq{repoColumnUserArticle.UserID: userID},
		sq.Eq{repoColumnUserArticle.ArticleID: articleID},
	}

	query, args, err := r.pgsq.Update(repoTableUserArticle).
		SetMap(update).
		Where(where).
		ToSql()
	if err != nil {
		return common.NewError(common.ErrorCodeInternalProcess, errors.Wrap(err, "failed to build update query for user_article rate"))
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to update user_article rate"))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to get rows affected"))
	}

	if rowsAffected == 0 {
		return common.NewError(common.ErrorCodeResourceNotFound, errors.New("user article not found or rate is the same"))
	}

	return nil
}

func (r *PostgresRepository) RefreshMaterializedView(ctx context.Context) common.Error {
	_, err := r.db.ExecContext(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY materialized_articles_average_rate;")
	if err != nil {
		return common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to refresh materialized view"))
	}
	return nil
}

func (r *PostgresRepository) DeleteUserArticleRate(ctx context.Context, userID uuid.UUID, articleID uuid.UUID) common.Error {
	return r.UpdateUserArticleRate(ctx, userID, articleID, 0)
}

// --- metadata_fetch_retries table ---

type repoMetadataFetchRetry struct {
	ID            int64          `db:"id"`
	ArticleID     uuid.UUID      `db:"article_id"`
	URL           string         `db:"url"`
	RetryCount    int16          `db:"retry_count"`
	LastAttemptAt sql.NullTime   `db:"last_attempt_at"`
	NextAttemptAt sql.NullTime   `db:"next_attempt_at"`
	Status        int16          `db:"status"`
	ErrorMessage  sql.NullString `db:"error_message"`
}

func (r *repoMetadataFetchRetry) toDomain() *article.MetadataFetchRetry {
	var lastAttemptAt *time.Time
	if r.LastAttemptAt.Valid {
		lastAttemptAt = &r.LastAttemptAt.Time
	}
	var nextAttemptAt *time.Time
	if r.NextAttemptAt.Valid {
		nextAttemptAt = &r.NextAttemptAt.Time
	}

	return &article.MetadataFetchRetry{
		ID:            r.ID,
		ArticleID:     r.ArticleID,
		URL:           r.URL,
		RetryCount:    r.RetryCount,
		LastAttemptAt: lastAttemptAt,
		NextAttemptAt: nextAttemptAt,
		Status:        article.RetryStatus(r.Status),
		ErrorMessage:  r.ErrorMessage.String,
	}
}

const repoTableMetadataFetchRetries = "metadata_fetch_retries"

type repoColumnPatternMetadataFetchRetries struct {
	ID            string
	ArticleID     string
	URL           string
	RetryCount    string
	LastAttemptAt string
	NextAttemptAt string
	Status        string
	ErrorMessage  string
}

var repoColumnMetadataFetchRetries = repoColumnPatternMetadataFetchRetries{
	ID:            "id",
	ArticleID:     "article_id",
	URL:           "url",
	RetryCount:    "retry_count",
	LastAttemptAt: "last_attempt_at",
	NextAttemptAt: "next_attempt_at",
	Status:        "status",
	ErrorMessage:  "error_message",
}

func (c repoColumnPatternMetadataFetchRetries) columns() string {
	return strings.Join([]string{
		c.ID,
		c.ArticleID,
		c.URL,
		c.RetryCount,
		c.LastAttemptAt,
		c.NextAttemptAt,
		c.Status,
		c.ErrorMessage,
	}, ", ")
}

func (r *PostgresRepository) CreateMetadataFetchRetry(ctx context.Context, articleID uuid.UUID, url string) common.Error {
	insert := map[string]interface{}{
		repoColumnMetadataFetchRetries.ArticleID: articleID,
		repoColumnMetadataFetchRetries.URL:       url,
	}

	query, args, err := r.pgsq.Insert(repoTableMetadataFetchRetries).
		SetMap(insert).
		Suffix(fmt.Sprintf("ON CONFLICT (%s) DO NOTHING", repoColumnMetadataFetchRetries.ArticleID)).
		ToSql()
	if err != nil {
		return common.NewError(common.ErrorCodeInternalProcess, errors.Wrap(err, "failed to build insert query for metadata_fetch_retries"))
	}

	if _, err = r.db.ExecContext(ctx, query, args...); err != nil {
		return common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to insert metadata_fetch_retries"))
	}

	return nil
}

func (r *PostgresRepository) GetPendingMetadataFetchRetries(ctx context.Context) ([]*article.MetadataFetchRetry, common.Error) {
	query, args, err := r.pgsq.Select(repoColumnMetadataFetchRetries.columns()).
		From(repoTableMetadataFetchRetries).
		Where(sq.And{
			sq.Eq{repoColumnMetadataFetchRetries.Status: 0}, // Pending status
			sq.LtOrEq{repoColumnMetadataFetchRetries.RetryCount: 3},
			sq.Or{
				sq.Eq{repoColumnMetadataFetchRetries.NextAttemptAt: nil},
				sq.LtOrEq{repoColumnMetadataFetchRetries.NextAttemptAt: time.Now()},
			},
		}).
		ToSql()
	if err != nil {
		return nil, common.NewError(common.ErrorCodeInternalProcess, errors.Wrap(err, "failed to build select query for pending metadata fetch retries"))
	}

	var rows []repoMetadataFetchRetry
	if err = r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to select pending metadata fetch retries"))
	}

	retries := make([]*article.MetadataFetchRetry, 0, len(rows))
	for _, row := range rows {
		retries = append(retries, row.toDomain())
	}

	return retries, nil
}

func (r *PostgresRepository) UpdateMetadataFetchRetryStatus(ctx context.Context, retryID int64, status int16, errorMessage string) common.Error {
	update := map[string]interface{}{
		repoColumnMetadataFetchRetries.Status:        status,
		repoColumnMetadataFetchRetries.ErrorMessage:  errorMessage,
		repoColumnMetadataFetchRetries.LastAttemptAt: time.Now(),
	}

	query, args, err := r.pgsq.Update(repoTableMetadataFetchRetries).
		SetMap(update).
		Where(sq.Eq{repoColumnMetadataFetchRetries.ID: retryID}).
		ToSql()
	if err != nil {
		return common.NewError(common.ErrorCodeInternalProcess, errors.Wrap(err, "failed to build update query for metadata fetch retry status"))
	}

	if _, err = r.db.ExecContext(ctx, query, args...); err != nil {
		return common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to update metadata fetch retry status"))
	}

	return nil
}

func (r *PostgresRepository) IncrementMetadataFetchRetryCount(ctx context.Context, retryID int64) common.Error {
	update := map[string]interface{}{
		repoColumnMetadataFetchRetries.RetryCount:    sq.Expr(fmt.Sprintf("%s + 1", repoColumnMetadataFetchRetries.RetryCount)),
		repoColumnMetadataFetchRetries.LastAttemptAt: time.Now(),
		repoColumnMetadataFetchRetries.NextAttemptAt: sq.Expr(fmt.Sprintf("NOW() + INTERVAL '%d minutes'", 5)), // Next attempt in 5 minutes
	}

	query, args, err := r.pgsq.Update(repoTableMetadataFetchRetries).
		SetMap(update).
		Where(sq.Eq{repoColumnMetadataFetchRetries.ID: retryID}).
		ToSql()
	if err != nil {
		return common.NewError(common.ErrorCodeInternalProcess, errors.Wrap(err, "failed to build update query for metadata fetch retry count"))
	}

	if _, err = r.db.ExecContext(ctx, query, args...); err != nil {
		return common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to increment metadata fetch retry count"))
	}

	return nil
}

func (r *PostgresRepository) GetArticleByID(ctx context.Context, articleID uuid.UUID) (*article.Article, common.Error) {
	query, args, err := r.pgsq.Select(repoColumnArticle.columns()).
		From(repoTableArticle).
		Where(sq.Eq{repoColumnArticle.ID: articleID}).
		ToSql()
	if err != nil {
		return nil, common.NewError(common.ErrorCodeInternalProcess, errors.Wrap(err, "failed to build select query for article by ID"))
	}

	var row repoArticle
	if err = r.db.GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewError(common.ErrorCodeResourceNotFound, err)
		}
		return nil, common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to select article by ID"))
	}

	return row.toDomain(), nil
}

func (r *PostgresRepository) UpdateArticle(ctx context.Context, art *article.Article) common.Error {
	update := map[string]interface{}{
		repoColumnArticle.Title:       art.Title,
		repoColumnArticle.Description: art.Description,
		repoColumnArticle.ImageURL:    art.ImageURL,
		repoColumnArticle.Metadata:    art.Metadata,
	}

	query, args, err := r.pgsq.Update(repoTableArticle).
		SetMap(update).
		Where(sq.Eq{repoColumnArticle.ID: art.ID}).
		ToSql()
	if err != nil {
		return common.NewError(common.ErrorCodeInternalProcess, errors.Wrap(err, "failed to build update query for article"))
	}

	if _, err = r.db.ExecContext(ctx, query, args...); err != nil {
		return common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to update article"))
	}

	return nil
}

func (r *PostgresRepository) GetTopRatedArticlesExcludingUser(ctx context.Context, excludedUserID uuid.UUID, limit int) (article.RecommendationArticles, common.Error) {
	selectColumns := []string{
		fmt.Sprintf("%s.%s", repoTableArticle, repoColumnArticle.ID),
		fmt.Sprintf("%s.%s", repoTableArticle, repoColumnArticle.URL),
		fmt.Sprintf("%s.%s", repoTableArticle, repoColumnArticle.Title),
		fmt.Sprintf("%s.%s", repoTableArticle, repoColumnArticle.Description),
		fmt.Sprintf("%s.%s", repoTableArticle, repoColumnArticle.ImageURL),
		fmt.Sprintf("%s.%s", repoTableArticle, repoColumnArticle.Metadata),
		fmt.Sprintf("%s.%s", repoMaterializedViewArticleAverageRate, repoColumnArticle.AverageRating),
	}

	query, args, err := r.pgsq.Select(selectColumns...).
		From(repoMaterializedViewArticleAverageRate).
		Join(fmt.Sprintf("%s ON %s.%s = %s.%s",
			repoTableArticle,
			repoTableArticle, repoColumnArticle.ID,
			repoMaterializedViewArticleAverageRate, repoColumnArticle.ID)).
		LeftJoin(fmt.Sprintf("%s ON %s.%s = %s.%s AND %s.%s = ?",
			repoTableUserArticle,
			repoTableUserArticle, repoColumnUserArticle.ArticleID,
			repoTableArticle, repoColumnArticle.ID,
			repoTableUserArticle, repoColumnUserArticle.UserID), excludedUserID).
		Where(sq.Eq{fmt.Sprintf("%s.%s", repoTableUserArticle, repoColumnUserArticle.UserID): nil}).
		OrderBy(fmt.Sprintf("%s.%s DESC", repoMaterializedViewArticleAverageRate, repoColumnArticle.AverageRating)).
		Limit(uint64(limit)).
		ToSql()
	if err != nil {
		return nil, common.NewError(common.ErrorCodeInternalProcess, errors.Wrap(err, "failed to build select query for top rated articles excluding user"))
	}

	type repoRecommendation struct {
		repoArticle
		AverageRating float64 `db:"average_rating"`
	}

	var rows []repoRecommendation
	if err = r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, common.NewError(common.ErrorCodeRemoteProcess, errors.Wrap(err, "failed to select top rated articles excluding user"))
	}

	recommendations := make(article.RecommendationArticles, 0, len(rows))
	for _, row := range rows {
		recommendations = append(recommendations, article.RecommendationArticle{
			Article:       row.repoArticle.toDomain(),
			AverageRating: row.AverageRating,
		})
	}

	return recommendations, nil
}
