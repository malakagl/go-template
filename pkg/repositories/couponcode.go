package repositories

import (
	"context"

	"github.com/malakagl/kart-challenge/pkg/errors"
	"github.com/malakagl/kart-challenge/pkg/log"
	"github.com/malakagl/kart-challenge/pkg/models/db"
	"github.com/malakagl/kart-challenge/pkg/otel"
	"gorm.io/gorm"
)

type CouponCodeRepo struct {
	db *gorm.DB
}

func NewCouponCodeRepository(db *gorm.DB) CouponCodeRepo {
	return CouponCodeRepo{db: db}
}

func (r *CouponCodeRepo) CountFilesByCode(ctx context.Context, code string) (int64, error) {
	spanCtx, span := otel.Tracer(ctx, "couponCodeRepo.countFilesByCode")
	defer span.End()

	var count int64
	err := r.db.WithContext(spanCtx).Model(&db.CouponCode{}).
		Where("code = ?", code).
		Distinct("file_id").
		Count(&count).Error
	if err != nil {
		log.WithCtx(spanCtx).Error().Msgf("error counting coupon code: %v", err)
		span.RecordError(err)
		return 0, errors.ErrDatabaseError
	}

	return count, nil
}
