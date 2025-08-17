package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/sappy5678/DeeliAi/internal/domain/common"
	"github.com/sappy5678/DeeliAi/internal/domain/user"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *user.User) (*user.User, common.Error)
	GetUserByEmail(ctx context.Context, email string) (*user.User, common.Error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, common.Error)
}

type repoUser struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
}

const repoTableUser = "users"

type repoColumnPatternUser struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
}

var repoColumnUser = repoColumnPatternUser{
	ID:           "id",
	Email:        "email",
	Username:     "username",
	PasswordHash: "password_hash",
}

func (c *repoColumnPatternUser) columns() string {
	return strings.Join([]string{
		c.ID,
		c.Email,
		c.Username,
		c.PasswordHash,
	}, ", ")
}

func (r *PostgresRepository) CreateUser(ctx context.Context, param *user.User) (*user.User, common.Error) {
	insert := map[string]interface{}{
		repoColumnUser.Email:        param.Email,
		repoColumnUser.Username:     param.Username,
		repoColumnUser.PasswordHash: param.PasswordHash,
	}
	query, args, err := r.pgsq.Insert(repoTableUser).
		SetMap(insert).
		Suffix(fmt.Sprintf("returning %s", repoColumnUser.columns())).
		ToSql()
	if err != nil {
		return nil, common.NewError(common.ErrorCodeInternalProcess, err)
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		if _, cerr := r.GetUserByEmail(ctx, param.Email); cerr == nil {
			e := fmt.Errorf("user with email %s already exists", param.Email)
			return nil, common.NewError(common.ErrorCodeParameterInvalid, e, common.WithMsg(e.Error()))
		}
		r.logger(ctx).Error().Str("query", query).Err(err).Msg("failed to get user")
		return nil, common.NewError(common.ErrorCodeRemoteProcess, err)
	}

	result, cerr := r.GetUserByEmail(ctx, param.Email)
	if cerr != nil {
		return nil, cerr
	}

	return result, nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, common.Error) {
	query, args, err := r.pgsq.Select(repoColumnUser.columns()).
		From(repoTableUser).
		Where(sq.Eq{repoColumnUser.Email: email}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, common.NewError(common.ErrorCodeInternalProcess, err)
	}

	row := repoUser{}
	if err = r.db.GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewError(common.ErrorCodeResourceNotFound, err, common.WithMsg("user is not found"))
		}
		r.logger(ctx).Error().Str("query", query).Err(err).Msg("failed to get user by email")
		return nil, common.NewError(common.ErrorCodeRemoteProcess, err, common.WithMsg(query))
	}

	result := user.User(row)
	return &result, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, common.Error) {
	query, args, err := r.pgsq.Select(repoColumnUser.columns()).
		From(repoTableUser).
		Where(sq.Eq{repoColumnUser.ID: id}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, common.NewError(common.ErrorCodeInternalProcess, err)
	}
	row := repoUser{}

	if err = r.db.GetContext(ctx, &row, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.NewError(common.ErrorCodeResourceNotFound, err, common.WithMsg("user is not found"))
		}
		return nil, common.NewError(common.ErrorCodeRemoteProcess, err)
	}

	result := user.User(row)
	return &result, nil
}
