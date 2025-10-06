package cache

import (
	"awesomeProject/src/app/config"
	"awesomeProject/src/app/domain"
	"context"
	"strconv"
	"sync"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Elem struct {
	Video      *domain.Video
	Expiration time.Time
	Duration   time.Duration
}
type VideoCache struct {
	cache            map[string]*Elem
	logger           *zap.Logger
	stop             chan struct{}
	mutex            sync.RWMutex
	defaultTTL       time.Duration
	defaultFrequency time.Duration
}

func NewVideoCache(config *config.Config, log *zap.Logger) *VideoCache {
	cache := &VideoCache{
		cache:            make(map[string]*Elem),
		defaultTTL:       config.Cache.DefaultExpiration,
		defaultFrequency: config.Cache.DefaultFrequency,
		stop:             make(chan struct{}),
		logger:           log,
	}

	return cache
}

func (cache *VideoCache) StartGC() {
	for {
		select {
		case <-cache.stop:
			return
		case <-time.After(cache.defaultFrequency):
			cache.CleanExpired()
		}
	}
}

func (cache *VideoCache) CleanExpired() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	now := time.Now()
	var k int = 0
	for key, elem := range cache.cache {
		if !elem.Expiration.IsZero() && now.After(elem.Expiration) {
			cache.logger.Info("deleted from cache", zap.String("slug", elem.Video.Slug))
			delete(cache.cache, key)
			k++
		}
	}
	cache.logger.Info("cache cleared", zap.String("items deleted", strconv.Itoa(k)))
}

func (cache *VideoCache) Set(id string, video *domain.Video, ttl time.Duration) {
	if ttl == 0 {
		ttl = cache.defaultTTL
	}

	var expiration time.Time

	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.cache[id] = &Elem{
		Video:      video,
		Expiration: expiration,
		Duration:   ttl,
	}
}

func (cache *VideoCache) Stop() {
	close(cache.stop)
}

func (cache *VideoCache) Get(id string) (*domain.Video, bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	elem, ok := cache.cache[id]
	//for debugging
	{
		for _, elem := range cache.cache {
			cache.logger.Info("elem in cache", zap.String("slug", elem.Video.Slug), zap.String("elem expiration", elem.Expiration.String()), zap.String("elem duration", elem.Duration.String()))
		}
	}

	if !ok {
		return nil, false
	}

	if !elem.Expiration.IsZero() && time.Now().After(elem.Expiration) {
		return nil, false
	}
	return elem.Video, true
}

func (cache *VideoCache) Delete(id string) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	delete(cache.cache, id)
	cache.logger.Info("deleted from cache", zap.String("id", id))

}

var CacheModule = fx.Module("video-cache",
	fx.Provide(NewVideoCache),
	fx.Invoke(func(lc fx.Lifecycle, c *VideoCache) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				if c.defaultFrequency > 0 {
					go c.StartGC()
				}
				return nil
			},
			OnStop: func(ctx context.Context) error {
				c.Stop()
				return nil
			},
		})
	}),
)
