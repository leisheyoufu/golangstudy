package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

var (
	endpoint string
	topic    string
	jobId    string
)

func main() {
	flag.Parse()
	topics := []string{}
	var wg = &sync.WaitGroup{}
	wg.Add(1)
	//广播式消费：消费者1
	go clusterConsumer(wg, []string{endpoint}, topics, "abc")
	//广播式消费：消费者2
	//go clusterConsumer(wg, []string{endpoint}, topic, "group-2")

	wg.Wait()
}

// 支持brokers cluster的消费者
func clusterConsumer(wg *sync.WaitGroup, brokers, topics []string, groupId string) {
	defer wg.Done()
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Offsets.CommitInterval = 1 * time.Second

	// init consumer
	consumer, err := cluster.NewConsumer(brokers, groupId, topics, config)
	if err != nil {
		log.Printf("%s: sarama.NewSyncProducer err, message=%s \n", groupId, err)
		return
	}
	defer consumer.Close()

	// trap SIGINT to trigger a shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	// consume errors
	go func() {
		for err := range consumer.Errors() {
			log.Printf("%s:Error: %s\n", groupId, err.Error())
		}
	}()

	// consume notifications
	go func() {
		for ntf := range consumer.Notifications() {
			log.Printf("%s:Rebalanced: %+v \n", groupId, ntf)
		}
	}()

	// consume messages, watch signals
	var successes int
Loop:
	for {
		select {
		case msg, ok := <-consumer.Messages():
			if ok {
				for _, header := range msg.Headers {
					if string(header.Key) == jobId {
						fmt.Fprintf(os.Stdout, "jobId %s %s\n", jobId, msg.Value)
					}
					//fmt.Fprintf(os.Stdout, "%s:%s/%d/%d\t%s\t%s\n", groupId, msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
					//consumer.MarkOffset(msg, "") // mark message as processed
					successes++
					fmt.Fprintf(os.Stdout, "%s:%s/%d/%d\t%s\t%s\n", groupId, msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
				}
			}
		case <-signals:
			break Loop
		}
	}
	fmt.Fprintf(os.Stdout, "%s consume %d messages \n", groupId, successes)

}

func init() {
	flag.StringVar(&endpoint, "endpoint", "", "Endpoint for kafka, format: 192.168.126.151:9092")
	flag.StringVar(&topic, "topic", "", "topic for kafka")
	flag.StringVar(&jobId, "jobId", "", "jobId kafka")
}
