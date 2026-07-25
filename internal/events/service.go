package events

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, event EventLog) (EventLog, error) {
	return s.repo.Create(ctx, event)
}

func (s *Service) ListByUser(ctx context.Context, userID string) ([]EventLog, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Summary(ctx context.Context, filter ReportFilter) ([]SummaryRow, error) {
	return s.repo.Summary(ctx, filter)
}

func (s *Service) ListFiltered(ctx context.Context, filter ReportFilter) ([]EventLog, error) {
	return s.repo.ListFiltered(ctx, filter)
}
