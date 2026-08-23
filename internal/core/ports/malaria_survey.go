package ports

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
)

type MalariaSurveyRepository interface {
	Create(ctx context.Context, survey *domain.MalariaSurvey) error
	GetPaginatedHistory(ctx context.Context, userID string, page, limit int) ([]domain.MalariaSurvey, int64, error)
	Update(ctx context.Context, id string, survey *domain.MalariaSurvey) error
}

type MalariaSurveyService interface {
	CreateSurvey(ctx context.Context, survey *domain.MalariaSurvey) error
	GetPaginatedHistory(ctx context.Context, userID string, page, limit int) ([]domain.MalariaSurvey, int64, error)
	UpdateSurvey(ctx context.Context, id string, survey *domain.MalariaSurvey) error
}
