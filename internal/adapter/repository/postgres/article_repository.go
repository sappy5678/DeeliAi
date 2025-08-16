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

	"github.com/chatbotgang/go-clean-architecture-template/internal/domain/article"
	"github.com/chatbotgang/go-clean-architecture-template/internal/domain/common"
)

// --- article table ---

type repoArticle struct {
	ID          uuid.UUID       `db:"id"`
	URL         string          `db:"url"`
	Title       sql.NullString  `db:"title"`
	Description sql.NullString  `db:"description"`
	ImageURL    sql.NullString  `db:"image_url"`
	Metadata    json.RawMessage `db:"metadata"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

func (a *repoArticle) toDomain() *article.Article {
	return &article.Article{
		ID:          a.ID,
		URL:         a.URL,
		Title:       a.Title.String,
		Description: a.Description.String,
		ImageURL:    a.ImageURL.String,
		Metadata:    a.Metadata,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

const repoTableArticle = "articles"

type repoColumnPatternArticle struct {
	ID          string
	URL         string
	Title       string
	Description string
	ImageURL    string
	Metadata    string
	CreatedAt   string
	UpdatedAt   string
}

var repoColumnArticle = repoColumnPatternArticle{
	ID:          "id",
	URL:         "url",
	Title:       "title",
	Description: "description",
	ImageURL:    "image_url",
	Metadata:    "metadata",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
}

func (c repoColumnPatternArticle) columns() string {
	col := []string{
		c.ID,
		c.URL,
		c.Title,
		c.Description,
		c.ImageURL,
		c.Metadata,
		c.CreatedAt,
		c.UpdatedAt,
	}
	for i, v := range col {
		col[i] = fmt.Sprintf("%s.%s", repoTableArticle, v)
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

func (r *PostgresRepository) DeleteUserArticleRate(ctx context.Context, userID uuid.UUID, articleID uuid.UUID) common.Error {
	return r.UpdateUserArticleRate(ctx, userID, articleID, 0)
}
