package users

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrTelegramChatLinked = errors.New("telegram chat is already linked to another user")
)

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

func (s *Service) GetByTelegramChatID(ctx context.Context, chatID string) (User, error) {
	return s.repo.GetByTelegramChatID(ctx, chatID)
}

func (s *Service) LinkTelegramChat(ctx context.Context, userID, chatID string) (User, error) {
	if chatID == "" {
		return User{}, errors.New("telegram chat id is required")
	}
	if userID == "" {
		return User{}, errors.New("user id is required")
	}

	tx, err := s.repo.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, fmt.Errorf("begin telegram link: %w", err)
	}
	defer tx.Rollback()

	existing, err := s.repo.getByTelegramChatID(ctx, tx, chatID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return User{}, err
	}
	if existing.ID != "" && existing.ID != userID {
		return User{}, ErrTelegramChatLinked
	}

	user, err := s.repo.getByID(ctx, tx, userID)
	if err != nil {
		return User{}, err
	}

	user.TelegramChatID = chatID
	if _, err := s.repo.update(ctx, tx, user); err != nil {
		return User{}, fmt.Errorf("link telegram chat: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return User{}, fmt.Errorf("commit telegram link: %w", err)
	}

	return s.repo.GetByID(ctx, userID)
}

func (s *Service) Update(ctx context.Context, user User) (User, error) {
	return s.repo.Update(ctx, user)
}
