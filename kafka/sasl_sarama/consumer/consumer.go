package main

import (
	"crypto/sha512"
	"flag"
	"fmt"
	"hash"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/leisheyoufu/golangstudy/kafka/sasl_sarama"
	"github.com/xdg/scram"
)

var (
	interval  int // Millisecond
	endpoints string
	user      string
	password  string
	topic     string
	group     string
	active                           = true
	SHA512    scram.HashGeneratorFcn = func() hash.Hash { return sha512.New() }
	logger                           = log.New(os.Stdout, "[Producer] ", log.LstdFlags)
	before    int
	partition int
	offset    int64
)

func init() {
	flag.StringVar(&endpoints, "endpoints", "", "Endpoint for kafka, format: 192.168.126.151:9092")
	flag.StringVar(&user, "user", "admin", "kafka user")
	flag.StringVar(&password, "pass", "", "kafka user password")
	flag.StringVar(&topic, "topic", "demo", "kafka topic")
	flag.StringVar(&group, "group", "demo", "kafka topic")
	flag.IntVar(&interval, "interval", 1000, "sleep time when producing message")
	flag.IntVar(&before, "before", 10, "seconds before current time")
	flag.IntVar(&partition, "partition", 0, "kafka partition")
	flag.Int64Var(&offset, "offset", 0, "kafka offset")
	sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
}

func partitionConsumer(wg *sync.WaitGroup, brokers, topics []string) {
	defer wg.Done()
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	//config.Consumer.Offsets.Initial = 10
	config.Consumer.Offsets.CommitInterval = 1 * time.Second
	config.Consumer.MaxProcessingTime = 500 * time.Microsecond
	config.Consumer.MaxWaitTime = 1000 * time.Microsecond
	config.Metadata.Full = true
	config.Net.SASL.Enable = true
	config.Net.SASL.User = user
	config.Net.SASL.Password = password
	config.Net.SASL.Handshake = true
	config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &sasl_sarama.XDGSCRAMClient{HashGeneratorFcn: SHA512} }
	config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512

	//saramaClient, err := sarama.NewClient(brokers, config)
	//if err != nil {
	//	log.Printf("Failed to init client error: %v\n", err)
	//	return
	//}
	//t := time.Now().Add(time.Duration(0-before) * time.Second)
	//fmt.Printf("Unix nano time %v %d\n", t, t.UnixNano()/1000)
	//offset, err := saramaClient.GetOffset(topics[partition], int32(partition), t.UnixNano()/1000000)
	//if err != nil {
	//	log.Printf("Get offset error: %v\n", err)
	//	return
	//}
	//	fmt.Printf("Time: %v offset: %d\n", t, offset)
	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		log.Printf("Failed to init consumer error: %v\n", err)
		return
	}
	partitionConsumer, err := consumer.ConsumePartition(topics[partition], int32(partition), offset)
	if err != nil {
		log.Printf("Failed to init consumer error: %v\n", err)
		return
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	consumed := 0
Loop:
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			if len(msg.Value) < 1024 && len(msg.Headers) >= 2 {
				log.Printf("Consumed message offset %d, header=%s, value=%s\n", msg.Offset, string(msg.Headers[1].Value), string(msg.Value))
			} else if len(msg.Value) < 1024 {
				log.Printf("Consumed message offset %d, value=%s\n", msg.Offset, string(msg.Value))
			} else if len(msg.Headers) >= 2 {
				log.Printf("Consumed message offset %d, header=%s\n", msg.Offset, string(msg.Headers[1].Value))
			} else {
				log.Printf("Consumed message offset %d", msg.Offset)
			}
			consumed++
		case <-signals:
			break Loop
		}
	}
}

func clusterConsumer(wg *sync.WaitGroup, brokers, topics []string) {
	defer wg.Done()
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.ClientID = "consumer_sarama"
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	//config.Consumer.Offsets.Initial = 10
	config.Consumer.Offsets.CommitInterval = 1 * time.Second
	config.Metadata.Full = true
	config.Net.SASL.Enable = true
	config.Net.SASL.User = user
	config.Net.SASL.Password = password
	config.Net.SASL.Handshake = true
	config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &sasl_sarama.XDGSCRAMClient{HashGeneratorFcn: SHA512} }
	config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
	// init consumer
	consumer, err := cluster.NewConsumer(brokers, group, topics, config)
	if err != nil {
		log.Printf("%s: sarama.NewConsumer err, message=%s \n", group, err)
		return
	}
	defer consumer.Close()

	// trap SIGINT to trigger a shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// consume errors
	go func() {
		for err := range consumer.Errors() {
			log.Printf("%s:Error: %s\n", group, err.Error())
		}
	}()

	// consume notifications
	go func() {
		for ntf := range consumer.Notifications() {
			log.Printf("%s:Rebalanced: %+v \n", group, ntf)
		}
	}()

	// consume messages, watch signals
	var successes int
	now := time.Now()
Loop:
	for {
		select {
		case msg, ok := <-consumer.Messages():
			if ok {
				if time.Now().Sub(now).Seconds() > 3 {
					now = time.Now()
					fmt.Fprintf(os.Stdout, "%s:%s/%d/%d\n", group, msg.Topic, msg.Partition, msg.Offset)
				}
				time.Sleep(time.Duration(interval) * time.Microsecond)
				consumer.MarkOffset(msg, "") // mark message as processed
				successes++
			}
		case <-signals:
			fmt.Println("Received signal")
			break Loop
		}
	}
	fmt.Fprintf(os.Stdout, "%s consume %d messages \n", group, successes)
}

func main() {
	flag.Parse()
	var wg = &sync.WaitGroup{}
	topics := []string{topic}
	wg.Add(1)
	//广播式消费：消费者1
	//go clusterConsumer(wg, []string{endpoints}, topics, "console-consumer2")
	//广播式消费：消费者2
	//clusterConsumer(wg, []string{endpoints}, topics)
	partitionConsumer(wg, []string{endpoints}, topics)
	wg.Wait()
}
