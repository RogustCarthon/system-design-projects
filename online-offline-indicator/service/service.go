package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

type Service struct {
	cache *redis.Client
}

var ServerError = map[string]string{"message": "internal server error"}

func NewService(ctx context.Context) (*Service, error) {
	svc := &Service{}
	if r, err := NewRedisClient(); err == nil {
		svc.cache = r
	} else {
		return nil, fmt.Errorf("failed to create service. [%w]", err)
	}
	return svc, nil
}

func (s *Service) Heartbeat(c *gin.Context) {
	if err := s.cache.Set(c.Param("userId"), true, 10*time.Second).Err(); err == nil {
		c.Status(http.StatusOK)
	} else {
		c.JSON(http.StatusInternalServerError, ServerError)
	}
}

func (s *Service) GetStatus(c *gin.Context) {
	if exists, err := s.cache.Exists(c.Param("userId")).Result(); err == nil && exists == 1 {
		c.JSON(http.StatusOK, map[string]interface{}{"isOnline": true})
	} else if err == nil {
		c.JSON(http.StatusOK, map[string]interface{}{"isOnline": false})
	} else {
		c.JSON(http.StatusInternalServerError, ServerError)
	}
}
