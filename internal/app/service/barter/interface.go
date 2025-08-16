package barter

import (
	"context"

	"github.com/sappy5678/DeeliAi/internal/domain/barter"
	"github.com/sappy5678/DeeliAi/internal/domain/common"
)

//go:generate mockgen -destination automock/good_repository.go -package=automock . GoodRepository
type GoodRepository interface {
	CreateGood(ctx context.Context, param barter.Good) (*barter.Good, common.Error)
	GetGoodByID(ctx context.Context, id int) (*barter.Good, common.Error)
	ListGoods(ctx context.Context) ([]barter.Good, common.Error)
	ListGoodsByOwner(ctx context.Context, ownerID int) ([]barter.Good, common.Error)
	UpdateGoods(ctx context.Context, goods []barter.Good) ([]barter.Good, common.Error)
	DeleteGoodByID(ctx context.Context, id int) common.Error
}
