package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	gocache "github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

type userGroupRateResolver struct {
	repo         UserGroupRateRepository
	cache        *gocache.Cache
	cacheTTL     time.Duration
	sf           *singleflight.Group
	logComponent string
}

func newUserGroupRateResolver(repo UserGroupRateRepository, cache *gocache.Cache, cacheTTL time.Duration, sf *singleflight.Group, logComponent string) *userGroupRateResolver {
	if cacheTTL <= 0 {
		cacheTTL = defaultUserGroupRateCacheTTL
	}
	if cache == nil {
		cache = gocache.New(cacheTTL, time.Minute)
	}
	if logComponent == "" {
		logComponent = "service.gateway"
	}
	if sf == nil {
		sf = &singleflight.Group{}
	}

	return &userGroupRateResolver{
		repo:         repo,
		cache:        cache,
		cacheTTL:     cacheTTL,
		sf:           sf,
		logComponent: logComponent,
	}
}

func (r *userGroupRateResolver) Resolve(ctx context.Context, userID, groupID int64, groupDefaultMultiplier float64) UserGroupRateOverride {
	fallback := UserGroupRateOverride{RateMultiplier: groupDefaultMultiplier}
	if r == nil || userID <= 0 || groupID <= 0 {
		return fallback
	}

	key := fmt.Sprintf("%d:%d", userID, groupID)
	if r.cache != nil {
		if cached, ok := r.cache.Get(key); ok {
			if override, castOK := cached.(UserGroupRateOverride); castOK {
				userGroupRateCacheHitTotal.Add(1)
				return override
			}
		}
	}
	if r.repo == nil {
		return fallback
	}
	userGroupRateCacheMissTotal.Add(1)

	value, err, shared := r.sf.Do(key, func() (any, error) {
		if r.cache != nil {
			if cached, ok := r.cache.Get(key); ok {
				if override, castOK := cached.(UserGroupRateOverride); castOK {
					userGroupRateCacheHitTotal.Add(1)
					return override, nil
				}
			}
		}

		userGroupRateCacheLoadTotal.Add(1)
		userRate, repoErr := r.repo.GetByUserAndGroup(ctx, userID, groupID)
		if repoErr != nil {
			return nil, repoErr
		}

		result := fallback
		if userRate != nil {
			result.RateMultiplier = userRate.RateMultiplier
			result.PerRequestPrice = userRate.PerRequestPrice
		}
		if r.cache != nil {
			r.cache.Set(key, result, r.cacheTTL)
		}
		return result, nil
	})
	if shared {
		userGroupRateCacheSFSharedTotal.Add(1)
	}
	if err != nil {
		userGroupRateCacheFallbackTotal.Add(1)
		logger.LegacyPrintf(r.logComponent, "get user group rate failed, fallback to group default: user=%d group=%d err=%v", userID, groupID, err)
		return fallback
	}

	override, ok := value.(UserGroupRateOverride)
	if !ok {
		userGroupRateCacheFallbackTotal.Add(1)
		return fallback
	}
	return override
}
