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
}

type Adapter struct {
	cache *cache.Cache
	ttl   time.Duration
}

func NewCacheAdapter(ttl time.Duration, cache *cache.Cache) Cache {
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

type Stub struct {
	SetStub     func(ctx context.Context, key string, value interface{}) error
	GetStub     func(ctx context.Context, key string, dest interface{}) error
	RealAdapter Cache
}

func (cs Stub) Set(ctx context.Context, key string, value interface{}) error {
	if cs.SetStub != nil {
		return cs.SetStub(ctx, key, value)
	}
	if cs.RealAdapter != nil {
		return cs.RealAdapter.Set(ctx, key, value)
	}
	panic("No real adapter or stub defined for Set")
}

func (cs Stub) Get(ctx context.Context, key string, dest interface{}) error {
	if cs.GetStub != nil {
		return cs.GetStub(ctx, key, dest)
	}
	if cs.RealAdapter != nil {
		return cs.RealAdapter.Get(ctx, key, dest)
	}
	panic("No real adapter or stub defined for Get")
}
