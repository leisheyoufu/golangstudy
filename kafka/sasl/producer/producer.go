package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var (
	endpoints string
	user      string
	password  string
	topic     string
	active    = true
)

func init() {
	flag.StringVar(&endpoints, "endpoints", "", "Endpoint for kafka, format: 192.168.126.151:9092")
	flag.StringVar(&user, "user", "admin", "kafka user")
	flag.StringVar(&password, "pass", "", "kafka user password")
	flag.StringVar(&topic, "topic", "demo", "kafka topic")
}

func main() {
	flag.Parse()
	producerConfig := &kafka.ConfigMap{
		"bootstrap.servers": endpoints,
		"security.protocol": "SASL_PLAINTEXT",
		"sasl.mechanism":    "SCRAM-SHA-512",
		"sasl.username":     user,
		"sasl.password":     password}
	kp, err := kafka.NewProducer(producerConfig)
	if err != nil {
		log.Fatal("failed to create producer - ", err)
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	// producer goroutine
	go func() {
		fmt.Println("started producer goroutine")

		for active == true {
			select {
			case <-exit:
				active = false
			default:
				err = kp.Produce(&kafka.Message{TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: kafka.PartitionAny},
					Key:   []byte("key-" + time.Now().String()),
					Value: []byte("value-" + time.Now().String())}, nil)
				if err != nil {
					fmt.Println("failed to produce message", err)
				}
				time.Sleep(2 * time.Second)
			}
		}
	}()
	for active == true {
		select {
		case <-exit:
			active = false
		case e := <-kp.Events():
			if e == nil {
				continue
			}
			//fmt.Println(e)
			m := e.(*kafka.Message)
			if m.TopicPartition.Error != nil {
				fmt.Println("delivery failed ", m.TopicPartition.Error)
			}
			fmt.Println("delivered messaged", e)
		}
	}
	kp.Close()
	fmt.Println("closed producer")
}
