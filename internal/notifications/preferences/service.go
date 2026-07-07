package preferences

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Upsert(ctx context.Context, pref Preference) (Preference, error) {
	return s.repo.Upsert(ctx, pref)
}

func (s *Service) ListByUser(ctx context.Context, userID string) ([]Preference, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) CreatePause(ctx context.Context, pause Pause) (Pause, error) {
	return s.repo.CreatePause(ctx, pause)
}

func (s *Service) ListPausesByUser(ctx context.Context, userID string) ([]Pause, error) {
	return s.repo.ListPausesByUser(ctx, userID)
}
