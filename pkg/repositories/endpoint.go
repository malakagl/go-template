package repositories

import (
	"context"

	"github.com/malakagl/go-template/pkg/models/db"
	"github.com/malakagl/go-template/pkg/otel"
)

func (a *ApiKeyRepository) FindAllEndpoints(ctx context.Context) ([]db.Endpoint, error) {
	spanCtx, span := otel.Tracer(ctx, "apikeyRepo.findAllEndpoints")
	defer span.End()

	var endpoints []db.Endpoint
	if err := a.db.WithContext(spanCtx).Find(&endpoints).Error; err != nil {
		span.RecordError(err)
		return nil, err
	}

	return endpoints, nil
}
