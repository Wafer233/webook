package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/internal/domain"
)

type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
}
type RedisUserCache struct {
	//传 单机 redis
	client     redis.Cmdable
	expiration time.Duration
}

// 从外面注入
func NewUserCache(client redis.Cmdable) UserCache {
	return &RedisUserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

func (uc *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := uc.key(id)
	//如果数据不纯在 err = redis,nil
	val, err := uc.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	if err != nil {
		return domain.User{}, err
	}
	return u, err
}

func (uc *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := uc.key(u.Id)
	return uc.client.Set(ctx, key, val, uc.expiration).Err()
}

func (uc *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("id:info:%d", id)
}
