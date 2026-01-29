package http

import (
	"context"
	"fmt"

	"github.com/soyacen/grocer/internal/layout/internal/http/repository"
)

type Service struct {
	repo repository.Repository
}

func (s *Service) Run(ctx context.Context) error {
	fmt.Println("implement logic here")
	return nil
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo}
}
