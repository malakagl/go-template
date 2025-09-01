package repositories

import (
	"context"

	"github.com/malakagl/go-template/pkg/models/db"
	"github.com/malakagl/go-template/pkg/otel"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepo struct {
	db *gorm.DB
}

func NewOrderRepo(db *gorm.DB) OrderRepo {
	return OrderRepo{db: db}
}

// Create inserts a new order with products
func (r *OrderRepo) Create(ctx context.Context, order *db.Order) error {
	spanCtx, span := otel.Tracer(ctx, "orderRepo.create")
	defer span.End()

	if err := r.db.WithContext(spanCtx).Clauses(clause.Returning{}).Create(&order).Error; err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}
