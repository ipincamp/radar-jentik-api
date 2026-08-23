package ports

import (
	"context"

	"github.com/ipincamp/radar-jentik-api/internal/core/domain"
)

type MalariaSurveyRepository interface {
	Create(ctx context.Context, survey *domain.MalariaSurvey) error
}

type MalariaSurveyService interface {
	CreateSurvey(ctx context.Context, survey *domain.MalariaSurvey) error
}
