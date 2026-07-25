package jobs

import (
	"context"
	"fmt"
	"sort"
)

type Job struct {
	Name string
	Run  func(context.Context) error
}

type Service struct {
	jobs map[string]Job
}

func NewService(defs []Job) *Service {
	jobsByName := map[string]Job{}
	for _, job := range defs {
		jobsByName[job.Name] = job
	}
	return &Service{jobs: jobsByName}
}

func (s *Service) Names() []string {
	names := make([]string, 0, len(s.jobs))
	for name := range s.jobs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (s *Service) Run(ctx context.Context, name string) error {
	job, ok := s.jobs[name]
	if !ok {
		return fmt.Errorf("job %s is not registered", name)
	}
	if job.Run == nil {
		return fmt.Errorf("job %s has no runner", name)
	}
	return job.Run(ctx)
}
