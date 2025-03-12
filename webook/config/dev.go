//go:build !k8s

package config

var Config = WebookConfig{
	DB: DBConfig{
		DSN: "root:root@tcp(localhost:13306)/webook",
	},
	Redis: RedisConfig{
		Addr: "localhost:16379",
	},
}
