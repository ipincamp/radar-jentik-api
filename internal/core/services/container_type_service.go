package services

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
	"github.com/ipincamp/radar-jentik-api/internal/core/ports"
)

type containerTypeService struct {
	repo ports.ContainerTypeRepository
}

func NewContainerTypeService(repo ports.ContainerTypeRepository) ports.ContainerTypeService {
	return &containerTypeService{
		repo: repo,
	}
}

func (s *containerTypeService) GetActiveTypes(ctx context.Context) ([]domain.ContainerType, error) {
	return s.repo.FindAllActive(ctx)
}
