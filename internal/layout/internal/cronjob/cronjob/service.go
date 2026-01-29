package cronjob

import (
	"context"
	"fmt"
)

type Service struct {
	repo *Repository
}

func (s *Service) Run(ctx context.Context) error {
	fmt.Println("implement logic here")
	return nil
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}
