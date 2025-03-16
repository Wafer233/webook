package config

type WebookConfig struct {
	DB    DBConfig
	Redis RedisConfig
	Kafka KafkaConfig
}

type DBConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr string
}

type KafkaConfig struct {
	Addr []string
}
