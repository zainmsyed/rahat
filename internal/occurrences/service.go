package occurrences

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, occurrence Occurrence) (Occurrence, error) {
	return s.repo.Create(ctx, occurrence)
}

func (s *Service) SaveOpen(ctx context.Context, occurrence Occurrence) (Occurrence, error) {
	return s.repo.SaveOpen(ctx, occurrence)
}

func (s *Service) GetByID(ctx context.Context, id string) (Occurrence, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, occurrence Occurrence) (Occurrence, error) {
	return s.repo.Update(ctx, occurrence)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) ListByUser(ctx context.Context, userID string) ([]Occurrence, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) ListByTask(ctx context.Context, taskID string) ([]Occurrence, error) {
	return s.repo.ListByTask(ctx, taskID)
}
