package users

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, user User) (User, error) {
	return s.repo.Create(ctx, user)
}

func (s *Service) GetByID(ctx context.Context, id string) (User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByEmail(ctx context.Context, email string) (User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *Service) Update(ctx context.Context, user User) (User, error) {
	return s.repo.Update(ctx, user)
}
