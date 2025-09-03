package service

import (
	"context"
	"time"
	"wb-order-service/internal/cache"
	"wb-order-service/internal/model"
	"wb-order-service/internal/repo"
)

type Service struct {
	repo  repo.Repository
	cache cache.Cache
	cacheTTL time.Duration
}

func New(r repo.Repository, c cache.Cache, ttl time.Duration) *Service {
	return &Service{repo: r, cache: c, cacheTTL: ttl}
}

func (s *Service) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	if v, ok := s.cache.Get(id); ok { return v, nil }
	o, err := s.repo.GetOrder(ctx, id)
	if err != nil { return nil, err }
	s.cache.Set(id, o, s.cacheTTL)
	return o, nil
}

func (s *Service) UpsertOrder(ctx context.Context, o *model.Order) error {
	if err := s.repo.UpsertOrder(ctx, o); err != nil { return err }
	s.cache.Set(o.OrderUID, o, s.cacheTTL)
	return nil
}
