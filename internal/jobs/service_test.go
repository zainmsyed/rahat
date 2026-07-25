package jobs

import (
	"context"
	"testing"
)

func TestRunRegisteredJob(t *testing.T) {
	ran := false
	svc := NewService([]Job{{Name: "daily-schedule", Run: func(context.Context) error {
		ran = true
		return nil
	}}})
	if err := svc.Run(context.Background(), "daily-schedule"); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !ran {
		t.Fatal("expected job to run")
	}
}

func TestRunUnknownJob(t *testing.T) {
	svc := NewService(nil)
	if err := svc.Run(context.Background(), "missing"); err == nil {
		t.Fatal("expected error for missing job")
	}
}
