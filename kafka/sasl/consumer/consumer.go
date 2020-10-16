package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var (
	endpoints string
	user      string
	password  string
	topic     string
	active    = true
	group     string
)

func init() {
	flag.StringVar(&endpoints, "endpoints", "", "Endpoint for kafka, format: 192.168.126.151:9092")
	flag.StringVar(&user, "user", "admin", "kafka user")
	flag.StringVar(&password, "pass", "", "kafka user password")
	flag.StringVar(&topic, "topic", "demo", "kafka topic")
	flag.StringVar(&group, "group", "", "kafka topic")
}

func main() {
	flag.Parse()
	consumerConfig := &kafka.ConfigMap{
		"bootstrap.servers": endpoints,
		"security.protocol": "SASL_PLAINTEXT",
		"sasl.mechanism":    "SCRAM-SHA-512",
		"sasl.username":     user,
		"sasl.password":     password,
		"group.id":          group,
		"auto.offset.reset": "earliest",
	}
	kc, err := kafka.NewConsumer(consumerConfig)
	if err != nil {
		log.Fatal("failed to create consumer - ", err)
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	// producer goroutine
	fmt.Println("started consumer")

	err = kc.Subscribe(topic, nil)
	if err != nil {
		log.Fatalf("unable to subscribe to topic %s - %v", topic, err)
	}
	for active == true {
		select {
		case <-exit:
			active = false
		default:
			ke := kc.Poll(1000)
			if ke == nil {
				continue
			}

			switch e := ke.(type) {
			case *kafka.Message:
				fmt.Printf("received message from %s: %s\n",
					e.TopicPartition, string(e.Value))

			case kafka.Error:
				fmt.Fprintf(os.Stderr, "Error: %v: %v\n", e.Code(), e)
			}
		}
	}
	err = kc.Close()
	if err != nil {
		fmt.Println("failed to close consumer ", err)
		return
	}
	fmt.Println("closed consumer")
}
