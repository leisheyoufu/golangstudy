package main

import (
	"flag"
	"fmt"
	"github.com/pingcap/errors"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

const PartitionSeqKey = "ps"
const GlobalSeqKey = "gs"
const TimestampKey = "Timestamp"
const CheckpointKey = "Checkpoint"
const (
	MAX_INT64 = int64(math.MaxInt64)
	GB        = 1024 * 1024 * 1024
)

func MINInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

var (
	endpoint    string
	topic       string
	jobId       string
	debug       bool
	startOffset int64
	endOffset   int64
	partition   int
	timeStr     string
)

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

func consumeCheckpoints(endpoint, topic string, partitionOffset, partitionOldest map[int32]int64, searchTime time.Time,
	consumer sarama.Consumer, saramaClient sarama.Client) (map[int32]int64, error) {
	results := make(map[string]int64)
	count := 0
	minTimestamp := MAX_INT64
	searchTimestamp := searchTime.Unix()
	var timestamp int64
	var err error
	for {
		count = 0
		for partition, _ := range partitionOffset {
			key := fmt.Sprintf("%d_%d", partition, partitionOffset[partition])
			if _, exist := results[key]; !exist {
				timestamp, err = consumeCheckpoint(topic, partitionOffset[partition], partition, consumer)
				if err != nil {
					return nil, err
				}
				log.Printf("Topic %s, Partition %d, offset %d, timestamp %d", topic, partition, partitionOffset[partition], timestamp)
				if timestamp == 0 {
					results[key] = MAX_INT64
				} else {
					results[key] = timestamp
				}
			}
			minTimestamp = MINInt64(minTimestamp, results[key])
		}
		for partition, _ := range partitionOffset {
			key := fmt.Sprintf("%d_%d", partition, partitionOffset[partition])
			if results[key] == minTimestamp && minTimestamp != MAX_INT64 && minTimestamp < searchTimestamp || partitionOffset[partition] == partitionOldest[partition] {
				count++
			} else {
				partitionOffset[partition]--
			}
		}
		if count == len(partitionOffset) {
			for partition, _ := range partitionOffset {
				key := fmt.Sprintf("%d_%d", partition, partitionOffset[partition])
				log.Printf("Topic %s Key %s value %v", topic, key, results[key])
			}
			break
		}
	}
	if minTimestamp == MAX_INT64 {
		minTimestamp, err = consumePartition(topic, endpoint, partitionOffset[0], 0, saramaClient)
		if err != nil {
			log.Printf("Can not consume topic %s, partition %d, offset %d", topic, 0, partitionOffset[0])
			return partitionOffset, err
		}
	}
	return partitionOffset, nil
}

func consumeCheckpoint(topic string, offset int64, partition int32, consumer sarama.Consumer) (int64, error) {
	partitionConsumer, err := consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		return 0, errors.Errorf("Failed to init partition consumer for topic %s, partition %d, offset %d. Error: %v", topic, partition, offset, err)
	}
	defer partitionConsumer.Close()
	select {
	case msg := <-partitionConsumer.Messages():
		for _, header := range msg.Headers {
			if string(header.Key) == CheckpointKey {
				checkpointTime, err := strconv.ParseInt(strings.Split(string(header.Value), ":")[1], 10, 64)
				if err != nil {
					return 0, errors.Errorf("Topic %s, Invalid checkpoint format. Error: %v", topic, err)
				}
				return checkpointTime, nil
			}
		}
		return 0, nil
	case <-time.After(10 * time.Second):
		return 0, errors.Errorf("Timeout while fetching message from topic %s, partition %d, offset %d\n", topic, partition, offset)
	}
	return 0, errors.Errorf("Internal error")
}

func consumePartition(endpoint, topic string, offset int64, partition int32, saramaClient sarama.Client) (int64, error) {
	consumer, err := sarama.NewConsumer([]string{endpoint}, saramaClient.Config())
	if err != nil {
		return 0, errors.Errorf("Can not init kafka consumer. Error: %v", err)
	}
	defer consumer.Close()
	config := saramaClient.Config()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = false
	oldest, err := saramaClient.GetOffset(topic, partition, sarama.OffsetOldest)
	if err != nil {
		return 0, err
	}
	newest, err := saramaClient.GetOffset(topic, partition, sarama.OffsetNewest)
	if err != nil {
		return 0, err
	}
	if newest == oldest {
		return 0, errors.Errorf("newest offset %d, oldest offset %d, no message to consume", newest, oldest)
	}
	if offset > oldest {
		offset = offset - 1
	}
	if offset < oldest {
		offset = oldest
	}
	partitionConsumer, err := consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		return 0, errors.Errorf("Failed to init partition consumer for topic %s, partition %d, offset %d. Error: %v", topic, partition, offset, err)
	}
	defer partitionConsumer.Close()
	var timestamp int64
	select {
	case msg := <-partitionConsumer.Messages():
		found := false
		for _, header := range msg.Headers {
			if string(header.Key) == TimestampKey {
				found = true
				timestamp, err = strconv.ParseInt(string(header.Value), 10, 64)
				if err != nil {
					return 0, errors.Errorf("Timestamp %s format error", string(header.Value))
				}
			}
		}
		if !found {
			return 0, errors.Errorf("Message at topic %s, partition %d offset %d do not contain timestamp header", topic, partition, offset)
		}
	case <-time.After(10 * time.Second):
		return 0, errors.Errorf("Timeout while fetching message from topic %s, partition %d, offset %d\n", topic, partition, offset)
	}
	return timestamp, nil
}

func GetOffsetForTimestamp(saramaClient sarama.Client, endpoint, topic string, timestamp time.Time) (map[int32]int64, error) {
	partitions, err := saramaClient.Partitions(topic)
	if err != nil {
		return nil, err
	}
	partitionOffsets := make(map[int32]int64)
	partitionOldest := make(map[int32]int64)
	var outerError error
	wg := sync.WaitGroup{}
	var lock sync.Mutex
	wg.Add(len(partitions))
	for _, partition := range partitions {
		go func(partition int32) {
			defer wg.Done()
			offset, err := saramaClient.GetOffset(topic, partition, timestamp.UnixNano()/1000000)
			if err != nil {
				outerError = errors.Errorf("Failed to get kafka offset for timestamp %v. Error: %v", timestamp, err)
				log.Print(outerError)
				return
			}
			log.Printf("Topic %s, partition %d, timestamp %v, offset %d", topic, partition, timestamp, offset)
			oldest, err := saramaClient.GetOffset(topic, partition, sarama.OffsetOldest)
			if err != nil {
				outerError = errors.Errorf("Failed to get kafka oldest offset for timestamp %v. Error: %v", timestamp, err)
				log.Print(outerError)
				return
			}
			lock.Lock()
			partitionOldest[partition] = oldest
			lock.Unlock()
			newest, err := saramaClient.GetOffset(topic, partition, sarama.OffsetNewest)
			if err != nil {
				outerError = errors.Errorf("Failed to get kafka newest offset for topic %s, partition %d, timestamp %v. Error: %v", topic, partition, timestamp, err)
				log.Print(outerError)
				return
			}
			if offset <= oldest || offset == -1 {
				offset, err = getRecentOffset(endpoint, topic, partition, oldest, newest, timestamp, saramaClient)
				if err != nil {
					outerError = errors.Errorf("Can not get recent offset for topic %s partition %d, oldest %d, newest %d, timestamp %v. Error: %v", topic, partition, oldest, newest, timestamp, err)
					log.Print(outerError)
					return
				}
				if offset != -1 {
					log.Printf("Offset provided for topic %s partition %d is outside the range, use %d", topic, partition, offset)
				} else {
					offset = oldest + 1
					log.Printf("Offset provided for topic %s partition %d is outside the range, use the oldest %d", topic, partition, offset)
				}
			}
			lock.Lock()
			partitionOffsets[partition] = offset
			lock.Unlock()
		}(partition)
	}
	wg.Wait()
	if outerError != nil {
		return nil, outerError
	}
	consumer, err := sarama.NewConsumer([]string{endpoint}, saramaClient.Config())
	if err != nil {
		return nil, errors.Errorf("Can not init kafka consumer. Error: %v", err)
	}
	defer consumer.Close()
	return consumeCheckpoints(endpoint, topic, partitionOffsets, partitionOldest, timestamp, consumer, saramaClient)
}

func getRecentOffset(endpoint, topic string, partition int32, oldest, newest int64, timestamp time.Time, saramaClient sarama.Client) (int64, error) {
	newestTimestamp, err := consumePartition(topic, endpoint, newest, partition, saramaClient)
	if err != nil {
		return 0, err
	}
	if timestamp.Unix() > newestTimestamp {
		return newest - 1, nil
	}
	oldestTimestamp, err := consumePartition(topic, endpoint, oldest, partition, saramaClient)
	if err != nil {
		return 0, err
	}
	if timestamp.Unix() <= oldestTimestamp {
		return oldest, nil
	}
	return -1, nil
}

func init() {
	flag.StringVar(&endpoint, "endpoint", "", "Endpoint for kafka, format: 192.168.126.151:9092")
	flag.StringVar(&topic, "topic", "", "topic for kafka")
	flag.StringVar(&jobId, "jobId", "", "jobId kafka")
	flag.BoolVar(&debug, "debug", false, "debug")
	flag.Int64Var(&startOffset, "start", 0, "offset")
	flag.Int64Var(&endOffset, "end", 0, "end")
	flag.IntVar(&partition, "partition", 0, "partition")
	flag.StringVar(&timeStr, "time", "", "timestamp 2021-02-09 10:47:00")
}

func main() {
	flag.Parse()
	if startOffset == 0 {
		startOffset = sarama.OffsetOldest
	}
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Partitioner = sarama.NewManualPartitioner
	// to avoid data lost
	config.Producer.RequiredAcks = sarama.WaitForAll
	//config.Producer.Flush.Frequency = 5 * time.Second
	//config.Producer.Flush.MaxMessages = 2048

	config.Version = sarama.V0_11_0_0

	config.Net.MaxOpenRequests = 1 //如果大于1，单个batch失败重试的时候，单个partition有可能乱序

	// retry 60 * 1 second for producer
	config.Producer.Retry.Max = 60
	config.Producer.Retry.Backoff = 1 * time.Second

	// retry 60 * 1 second for metadata update
	config.Metadata.Retry.Max = 60
	config.Metadata.Retry.Backoff = 1 * time.Second

	config.Producer.MaxMessageBytes = 1048576
	config.Net.SASL.Enable = false
	config.Net.SASL.Handshake = false
	saramaClient, err := sarama.NewClient([]string{endpoint}, config)
	if err != nil {
		panic(err)
	}
	var t time.Time
	if timeStr == "" {
		t = time.Now()
	} else {
		var LOC, _ = time.LoadLocation("Asia/Shanghai")
		t, err = time.ParseInLocation("2006-01-02 15:04:05", timeStr, LOC)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid time, error %v", err)
			os.Exit(1)
		}
	}
	m, err := GetOffsetForTimestamp(saramaClient, endpoint, topic, t)
	fmt.Println(m)
}

// kafka-partition-consume --endpoint 100.119.167.50:7745 --topic progress --jobId ixyau9qa
