package ioc

import (
	"github.com/IBM/sarama"
	"log"
	"webook/config"
	"webook/internal/event"
)

func InitSaramaClient() sarama.Client {
	addrs := config.Config.Kafka.Addr

	scfg := sarama.NewConfig()
	scfg.Producer.Return.Successes = true

	client, err := sarama.NewClient(addrs, scfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka client: %v", err)
	}
	return client
}

func InitSyncProducer(c sarama.Client) sarama.SyncProducer {
	p, err := sarama.NewSyncProducerFromClient(c)
	if err != nil {
		panic(err)
	}
	return p
}

func InitConsumers(c1 *event.InteractiveReadEventConsumer) []event.Consumer {
	return []event.Consumer{c1}
}
