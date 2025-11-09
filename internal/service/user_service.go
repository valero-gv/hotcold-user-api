package service

import (
	"context"
	"fmt"
	"time"

	"azret/internal/cache"
	"azret/internal/model"
	"azret/internal/repository"

	"golang.org/x/sync/singleflight"
)

type UserService struct {
	repo       *repository.UserRepository
	cache      *cache.UserCache
	reqTimeout time.Duration
	sfGroup    singleflight.Group
}

func NewUserService(repo *repository.UserRepository, cache *cache.UserCache, reqTimeout time.Duration) *UserService {
	return &UserService{repo: repo, cache: cache, reqTimeout: reqTimeout}
}

// GetUser implements read-through cache with hot/cold paths.
// Hot (Redis): fastest path; if cache fails, we degrade gracefully to the cold path.
// Cold (PostgreSQL): authoritative primary-key lookup.
// Returns (user, cacheHit, found, err).
func (s *UserService) GetUser(ctx context.Context, id string) (model.User, bool, bool, error) {
	// Apply per-request timeout budget for cold path
	var cancel context.CancelFunc
	if s.reqTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, s.reqTimeout)
		defer cancel()
	}

	if u, ok, err := s.cache.Get(ctx, id); err == nil && ok {
		return u, true, true, nil
	} else if err != nil {
		// Cache error: fall through to DB (do not fail the request)
		_ = err
	}

	// Deduplicate concurrent cold lookups for the same user_id.
	type result struct {
		u     model.User
		found bool
		err   error
	}
	v, _, _ := s.sfGroup.Do("user:"+id, func() (any, error) {
		u, found, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return result{err: fmt.Errorf("repository GetByID user_id=%s: %w", id, err)}, nil
		}
		if !found {
			return result{found: false}, nil
		}
		// Best-effort cache set; ignore error
		_ = s.cache.Set(context.Background(), u)
		return result{u: u, found: true}, nil
	})
	res := v.(result)
	if res.err != nil {
		return model.User{}, false, false, res.err
	}
	if !res.found {
		return model.User{}, false, false, nil
	}
	return res.u, false, true, nil
}
