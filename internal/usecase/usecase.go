package usecase

import (
	"go_test/internal/domain"
)

type ProductUseCase struct {
	repo domain.ProductRepository
}

func NewProductUseCase(repo domain.ProductRepository) *ProductUseCase {
	return &ProductUseCase{
		repo: repo,
	}
}

func (puc *ProductUseCase) FetchProductInfo(orderIDs string) ([]domain.ProductInfo, error) {
	return puc.repo.GetProductInfo(orderIDs)
}
