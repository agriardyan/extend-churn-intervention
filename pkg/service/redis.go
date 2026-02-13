package service

import "github.com/go-redis/redis/v8"

type RedisService struct {
	client redis.UniversalClient
	cfg    RedisServiceConfig
}

type RedisServiceConfig struct{}

func NewRedisService(
	client redis.UniversalClient,
	cfg RedisServiceConfig,
) (*RedisService, error) {
	return &RedisService{
		client: client,
		cfg:    cfg,
	}, nil
}
