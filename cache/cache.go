package cache

import (
	"context"
	"errors"
	"github.com/go-redis/cache/v9"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
}

type Adapter struct {
	cache *cache.Cache
	ttl   time.Duration
}

func NewCacheAdapter(ttl time.Duration, cache *cache.Cache) *Adapter {
	return &Adapter{ttl: ttl, cache: cache}
}

func (adapter *Adapter) Set(ctx context.Context, key string, value interface{}) error {
	return adapter.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: value,
		TTL:   adapter.ttl,
	})
}

func (adapter *Adapter) Get(ctx context.Context, key string, dest interface{}) error {
	err := adapter.cache.Get(ctx, key, dest)
	if errors.Is(err, cache.ErrCacheMiss) {
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

func (adapter *Adapter) Delete(ctx context.Context, key string) error {
	//TODO implement me
	panic("implement me")
}
