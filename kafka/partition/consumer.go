package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"github.com/Shopify/sarama"
)

const PartitionSeqKey = "ps"
const GlobalSeqKey = "gs"
const TimestampKey = "Timestamp"
const CheckpointKey = "Checkpoint"

var (
	endpoint    string
	topic       string
	jobId       string
	debug       bool
	startOffset int64
	endOffset   int64
	partition   int
)

func main() {
	flag.Parse()
	if startOffset == 0 {
		startOffset = sarama.OffsetOldest
	}
	PartitionConsumeFromOffset(topic, int32(partition), startOffset)
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

func IsCheckpointMsg(msg *sarama.ConsumerMessage) bool {
	for _, m := range msg.Headers {
		if string(m.Key) == CheckpointKey {
			return true
		}
	}

	return false
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
			var currentJobId string
			for _, header := range msg.Headers {
				if string(header.Key) == "jobId" && string(header.Value) == jobId {
					fmt.Fprintf(os.Stdout, "offset %d jobid %s time: %v %s\n", msg.Offset, jobId, msg.Timestamp, msg.Value)
				}
				if string(header.Key) == "jobId" {
					currentJobId = string(header.Value)
				}
				//fmt.Fprintf(os.Stdout, "offset %d jobid %s %s\n", msg.Offset, jobId, header.Value)
			}
			gs := GetGlobalSeq(msg)
			ps := GetPartitionSeq(msg)
			isCkp := IsCheckpointMsg(msg)
			if msg.Offset > endOffset {
				break Loop
			}
			fmt.Fprintf(os.Stdout, "offset %d key: %s, jobid %s time: %v, gs: %d, ps: %d, ckp: %v\n", msg.Offset, string(msg.Key), jobId, msg.Timestamp, gs, ps, isCkp)

			if debug {
				fmt.Fprintf(os.Stdout, "%s: %s/%d/%d\t%v\t%s\t%s\n", currentJobId, msg.Topic, msg.Partition, msg.Offset, msg.Timestamp, msg.Key, msg.Value)
			}
		case <-signals:
			break Loop
		}
	}
	return nil
}

func init() {
	flag.StringVar(&endpoint, "endpoint", "", "Endpoint for kafka, format: 192.168.126.151:9092")
	flag.StringVar(&topic, "topic", "", "topic for kafka")
	flag.StringVar(&jobId, "jobId", "", "jobId kafka")
	flag.BoolVar(&debug, "debug", false, "debug")
	flag.Int64Var(&startOffset, "start", 0, "offset")
	flag.Int64Var(&endOffset, "end", 0, "end")
	flag.IntVar(&partition, "partition", 0, "partition")
}

// kafka-partition-consume --endpoint 100.119.167.50:7745 --topic progress --jobId ixyau9qa
