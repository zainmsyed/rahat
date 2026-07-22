package telegram

import (
	"context"
	"log/slog"
	"time"
)

type Poller struct {
	client         RuntimeClient
	handler        CallbackHandler
	logger         *slog.Logger
	timeoutSeconds int
	retryDelay     time.Duration
}

func NewPoller(client RuntimeClient, handler CallbackHandler, logger *slog.Logger) *Poller {
	if logger == nil {
		logger = slog.Default()
	}
	return &Poller{
		client:         client,
		handler:        handler,
		logger:         logger,
		timeoutSeconds: 30,
		retryDelay:     time.Second,
	}
}

func (p *Poller) Run(ctx context.Context) {
	var offset int64
	for {
		if ctx.Err() != nil {
			return
		}
		updates, err := p.client.GetUpdates(ctx, GetUpdatesRequest{Offset: offset, Timeout: p.timeoutSeconds})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			p.logger.Warn("telegram long polling failed", "error", err)
			if !sleepContext(ctx, p.retryDelay) {
				return
			}
			continue
		}

		processedAll := true
		for _, update := range updates {
			if err := p.handleUpdate(ctx, update); err != nil {
				processedAll = false
				p.logger.Warn("telegram update handling failed", "update_id", update.UpdateID, "error", err)
				break
			}
			offset = update.UpdateID + 1
		}
		if !processedAll && !sleepContext(ctx, p.retryDelay) {
			return
		}
	}
}

func (p *Poller) handleUpdate(ctx context.Context, update Update) error {
	if update.CallbackQuery != nil && update.CallbackQuery.Data != "" {
		return p.handler.HandleCallback(ctx, update.CallbackQuery.Data)
	}
	return nil
}

func sleepContext(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
