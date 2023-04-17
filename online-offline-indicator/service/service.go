package service

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type mp = map[string]interface{}

type Service struct {
	cache *redis.Client
}

var ServerError = map[string]string{"message": "internal server error"}

func NewService() (*Service, error) {
	svc := &Service{}
	if r, err := NewRedisClient(); err == nil {
		svc.cache = r
	} else {
		return nil, fmt.Errorf("failed to create service. [%w]", err)
	}
	return svc, nil
}

func (s *Service) Heartbeat(userId string) error {
	if err := s.cache.Set(userId, true, time.Second).Err(); err == nil {
		return nil
	} else {
		return fmt.Errorf("failed to set in cache. [%w]", err)
	}
}

func (s *Service) GetStatus(userId string) (bool, error) {
	if exists, err := s.cache.Exists(userId).Result(); err == nil && exists == 1 {
		return true, nil
	} else if err == nil {
		return false, nil
	} else {
		return false, fmt.Errorf("failed to check in cache. [%w]", err)
	}
}

func (s *Service) GetOnlineCount() (int, error) {
	if keys, err := s.cache.Keys("*").Result(); err == nil {
		return len(keys), nil
	} else {
		return 0, fmt.Errorf("failed to get keys from cache. [%w]", err)
	}
}
