package startup

import (
	"github.com/redis/go-redis/v9"
	"webook/config"
)

func InitRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		//Addr: "webook-redis:6381",
		Addr: config.Config.Redis.Addr,
	})

	return redisClient
}
