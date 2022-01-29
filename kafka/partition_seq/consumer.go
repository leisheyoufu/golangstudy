package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"github.com/Shopify/sarama"
)

var (
	endpoint string
	topic    string
	jobId    string
	debug    bool
	offset   int64
)

const PartitionSeqKey = "ps"
const GlobalSeqKey = "gs"

func main() {
	flag.Parse()
	if offset == 0 {
		offset = sarama.OffsetOldest
	}
	PartitionConsumeFromOffset(topic, 0, offset)
}

// 支持brokers cluster的消费者
func PartitionConsumeFromOffset(topic string, partition int32, offset int64) error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = false
	config.Version = sarama.V1_1_1_0
	//config.Metadata.Full = true
	consumer, err := sarama.NewConsumer([]string{endpoint}, config)
	if err != nil {
		fmt.Printf("Can not init kafka consumer. Error: %v", err)
		return err
	}
	defer consumer.Close()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	partitionConsumer, err := consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		fmt.Printf("Failed to init partition consumer for topic %s, partition %d, offset %d. Error: %v", topic, partition, offset, err)
		return err
	}
	defer partitionConsumer.Close()

Loop:
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			ps := GetPartitionSeq(msg)
			gs := GetGlobalSeq(msg)
			fmt.Printf("offset %d, ps: %d, gs: %d\n", msg.Offset, ps, gs)
		case <-signals:
			break Loop
		}
	}
	return nil
}

func getSeq(msg *sarama.ConsumerMessage, key string) uint64 {
	for _, m := range msg.Headers {
		if string(m.Key) == key {
			n, err := strconv.ParseUint(string(m.Value), 10, 64)
			if nil != err {
				panic("unexpected msg " + key + " seq: " + string(m.Value))
			}

			return n
		}
	}

	panic(key + " does not exists")
}

func GetPartitionSeq(msg *sarama.ConsumerMessage) uint64 {
	return getSeq(msg, PartitionSeqKey)
}

func GetGlobalSeq(msg *sarama.ConsumerMessage) uint64 {
	return getSeq(msg, GlobalSeqKey)
}

func init() {
	flag.StringVar(&endpoint, "endpoint", "", "Endpoint for kafka, format: 192.168.126.151:9092")
	flag.StringVar(&topic, "topic", "", "topic for kafka")
	flag.BoolVar(&debug, "debug", false, "debug")
	flag.Int64Var(&offset, "offset", 0, "offset")
}

// kafka-partition-consume --endpoint 100.119.167.50:7745 --topic progress --jobId ixyau9qa
