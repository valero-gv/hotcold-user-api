package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"azret/internal/model"
)

type UserCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewUserCache(client *redis.Client, ttl time.Duration) *UserCache {
	return &UserCache{client: client, ttl: ttl}
}

func userKey(id string) string { return fmt.Sprintf("user:%s", id) }

// Get returns (user, true, nil) if present; (_, false, nil) if not.
func (c *UserCache) Get(ctx context.Context, id string) (model.User, bool, error) {
	val, err := c.client.Get(ctx, userKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return model.User{}, false, nil
		}
		return model.User{}, false, err
	}
	var u model.User
	if err := json.Unmarshal(val, &u); err != nil {
		return model.User{}, false, err
	}
	return u, true, nil
}

// Set stores the user with configured TTL. If ttl == 0, key is persistent.
func (c *UserCache) Set(ctx context.Context, u model.User) error {
	b, err := json.Marshal(u)
	if err != nil {
		return err
	}
	if c.ttl > 0 {
		return c.client.Set(ctx, userKey(u.UserID), b, c.ttl).Err()
	}
	return c.client.Set(ctx, userKey(u.UserID), b, 0).Err()
}
