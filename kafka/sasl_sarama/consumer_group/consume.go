package main

import (
	"context"
	"crypto/sha512"
	"flag"
	"hash"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/leisheyoufu/golangstudy/kafka/sasl_sarama"
	"github.com/xdg/scram"
)

// Sarama configuration options
var (
	endpoints   = ""
	version     = ""
	group       = ""
	topic       = ""
	oldest      = true
	verbose     = false
	user        string
	password    string
	initOffsets = map[int32]int64{
		0: 9901,
		1: 9906,
		2: 9901,
		3: 9903,
	}
	SHA512 scram.HashGeneratorFcn = func() hash.Hash { return sha512.New() }
)

func init() {
	flag.StringVar(&endpoints, "endpoints", "", "Kafka bootstrap brokers to connect to, as a comma separated list")
	flag.StringVar(&group, "group", "admin", "Kafka consumer group definition")
	flag.StringVar(&version, "version", "2.6.0", "Kafka cluster version")
	flag.StringVar(&topic, "topic", "", "Kafka topic to be consumed, as a comma separated list")
	flag.BoolVar(&oldest, "oldest", true, "Kafka consumer consume initial offset from oldest")
	flag.BoolVar(&verbose, "verbose", false, "Sarama logging")
	flag.StringVar(&password, "pass", "", "kafka user password")
	flag.StringVar(&user, "user", "", "kafka user name")
	flag.Parse()

	if len(endpoints) == 0 {
		panic("no Kafka bootstrap brokers defined, please set the -brokers flag")
	}

	if len(topic) == 0 {
		panic("no topics given to be consumed, please set the -topics flag")
	}

	if len(group) == 0 {
		panic("no Kafka consumer group defined, please set the -group flag")
	}
	if len(password) == 0 {
		panic("no Kafka password defined, please set the -pass flag")
	}
	if len(user) == 0 {
		panic("no Kafka user defined, please set the -user flag")
	}
}

func main() {
	log.Println("Starting a new Sarama consumer")

	if verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	version, err := sarama.ParseKafkaVersion(version)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	/**
	 * Construct a new Sarama configuration.
	 * The Kafka cluster version has to be defined before the consumer/producer is initialized.
	 */
	config := sarama.NewConfig()
	config.Version = version
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Metadata.Full = true
	config.Net.SASL.Enable = true
	config.Net.SASL.User = user
	config.Net.SASL.Password = password
	config.Net.SASL.Handshake = true
	config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &sasl_sarama.XDGSCRAMClient{HashGeneratorFcn: SHA512} }
	config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512

	/**
	 * Setup a new Sarama consumer group
	 */
	consumer := Consumer{
		ready: make(chan bool),
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(strings.Split(endpoints, ","), group, config)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := client.Consume(ctx, strings.Split(topic, ","), &consumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		log.Println("terminating: context cancelled")
	case <-sigterm:
		log.Println("terminating: via signal")
	}
	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sess sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	for partition, offset := range initOffsets {
		sess.ResetOffset(topic, partition, offset, "")
	}
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s, partition = %d", string(message.Value), message.Timestamp, message.Topic, message.Partition)
		session.MarkMessage(message, "")
	}

	return nil
}
