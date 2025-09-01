package services

import (
	"context"
	"strconv"

	"github.com/malakagl/go-template/pkg/errors"
	"github.com/malakagl/go-template/pkg/log"
	"github.com/malakagl/go-template/pkg/models/dto/response"
	"github.com/malakagl/go-template/pkg/repositories"
)

type IProductService interface {
	FindAll(ctx context.Context) (*response.ProductsResponse, error)
	FindByID(ctx context.Context, id uint) (*response.ProductResponse, error)
}

type ProductService struct {
	repo repositories.ProductRepo
}

func NewProductService(r repositories.ProductRepo) ProductService {
	return ProductService{repo: r}
}

func (s *ProductService) FindAll(ctx context.Context) (*response.ProductsResponse, error) {
	res, err := s.repo.FindAll(ctx)
	if err != nil {
		log.WithCtx(ctx).Error().Msgf("findAll failed with error: %v", err)
		return nil, errors.ErrProductNotFound
	}

	if res == nil {
		log.WithCtx(ctx).Error().Msg("findAll returned 0 elements")
		return nil, errors.ErrProductNotFound
	}

	products := make([]response.Product, len(res))
	for i, p := range res {
		products[i] = response.Product{
			ID:       strconv.FormatUint(uint64(p.ID), 10),
			Name:     p.Name,
			Price:    p.Price,
			Category: p.Category,
			Image: response.ProductImage{
				Thumbnail: p.Image.Thumbnail,
				Mobile:    p.Image.Mobile,
				Tablet:    p.Image.Tablet,
				Desktop:   p.Image.Desktop,
			},
		}
	}

	return &response.ProductsResponse{Products: products}, nil
}

func (s *ProductService) FindByID(ctx context.Context, id uint) (*response.ProductResponse, error) {
	res, err := s.repo.FindByID(ctx, id)
	if err != nil {
		log.WithCtx(ctx).Error().Msgf("findAll failed with error: %v", err)
		return nil, errors.ErrProductNotFound
	}

	if res == nil {
		log.WithCtx(ctx).Error().Msg("findAll returned 0 elements")
		return nil, errors.ErrProductNotFound
	}

	return &response.ProductResponse{
		ID:       strconv.FormatUint(uint64(res.ID), 10),
		Name:     res.Name,
		Price:    res.Price,
		Category: res.Category,
		Image: response.ProductImage{
			Thumbnail: res.Image.Thumbnail,
			Mobile:    res.Image.Mobile,
			Tablet:    res.Image.Tablet,
			Desktop:   res.Image.Desktop,
		},
	}, nil
}
