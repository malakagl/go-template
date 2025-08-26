package repositories

import (
	"context"

	"github.com/malakagl/kart-challenge/pkg/errors"
	"github.com/malakagl/kart-challenge/pkg/log"
	"github.com/malakagl/kart-challenge/pkg/models/db"
	"github.com/malakagl/kart-challenge/pkg/otel"
	"gorm.io/gorm"
)

type ProductRepo struct {
	db *gorm.DB
}

func NewProductRepo(db *gorm.DB) ProductRepo {
	return ProductRepo{db: db}
}

func (r *ProductRepo) FindAll(ctx context.Context) ([]db.Product, error) {
	spanCtx, span := otel.Tracer(ctx, "productRepo.findAll")
	defer span.End()

	var products []db.Product
	if err := r.db.WithContext(spanCtx).Preload("Image").Find(&products).Error; err != nil {
		span.RecordError(err)
		return nil, errors.ErrProductNotFound
	}

	return products, nil
}

func (r *ProductRepo) FindByID(ctx context.Context, id uint) (*db.Product, error) {
	spanCtx, span := otel.Tracer(ctx, "productRepo.findByID")
	defer span.End()

	var product db.Product
	if err := r.db.WithContext(spanCtx).Preload("Image").First(&product, "id = ?", id).Error; err != nil {
		log.WithCtx(spanCtx).Error().Msgf("Error fetching product with ID %d: %v", id, err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.ErrProductNotFound
		}

		span.RecordError(err)
		return nil, errors.ErrDatabaseError
	}

	return &product, nil
}
