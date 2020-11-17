package main

import (
	"crypto/sha512"
	"errors"
	"flag"
	"fmt"
	"hash"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/leisheyoufu/golangstudy/kafka/sasl_sarama"
	"github.com/xdg/scram"
)

var (
	interval         int // Ms
	endpoints        string
	user             string
	password         string
	topic            string
	SHA512           scram.HashGeneratorFcn = func() hash.Hash { return sha512.New() }
	logger                                  = log.New(os.Stdout, "[Producer] ", log.LstdFlags)
	active                                  = true
	ErrTopicNotExist                        = errors.New("This request is for a topic or partition that does not exist on this broker")
)

func init() {
	flag.StringVar(&endpoints, "endpoints", "", "Endpoint for kafka, format: 192.168.126.151:9092")
	flag.StringVar(&user, "user", "admin", "kafka user")
	flag.StringVar(&password, "pass", "", "kafka user password")
	flag.StringVar(&topic, "topic", "demo", "kafka topic")
	flag.IntVar(&interval, "interval", 1000, "sleep time when producing message")
	sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
}

func createTopic(brokers []string, config *sarama.Config) error {
	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		log.Fatal("Error while creating cluster admin: ", err.Error())
	}
	defer admin.Close()

	err = admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     2,
		ReplicationFactor: 2,
	}, false)
	if err != nil {
		log.Fatal("Error while creating topic: ", err.Error())
		return err
	}
	return nil
}

func describeTopic(brokers []string, config *sarama.Config) error {
	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		log.Fatal("Error while describing cluster admin: ", err.Error())
		return err
	}
	defer func() { _ = admin.Close() }()
	metadata, err := admin.DescribeTopics([]string{topic})
	if err != nil {
		log.Fatal("Error while describing cluster admin: ", err.Error())
		return err
	}
	// kafka error code https://cwiki.apache.org/confluence/display/KAFKA/A+Guide+To+The+Kafka+Protocol#AGuideToTheKafkaProtocol-ErrorCodes
	if metadata[0].Err == 3 {
		return ErrTopicNotExist
	}
	return nil
}

func produce(brokers []string, conf *sarama.Config) error {
	syncProducer, err := sarama.NewSyncProducer(brokers, conf)
	if err != nil {
		logger.Fatalln("failed to create producer: ", err)
		return err
	}
	defer syncProducer.Close()
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	now := time.Now()
	for active == true {
		select {
		case <-exit:
			active = false
		default:
			partition, offset, err := syncProducer.SendMessage(&sarama.ProducerMessage{
				Topic: topic,
				Value: sarama.StringEncoder("test_message-" + time.Now().String()),
				Headers: []sarama.RecordHeader{
					{Key: []byte("Name"), Value: []byte("golang")},
					{Key: []byte("Timestamp"), Value: []byte(strconv.FormatInt(time.Now().Unix(), 10))},
				},
			})
			if err != nil {
				logger.Fatalln("failed to send message to ", topic, err)
				return err
			}
			if time.Now().Sub(now).Seconds() > 3 {
				now = time.Now()
				logger.Printf("wrote message at partition: %d, offset: %d", partition, offset)
			}
			time.Sleep(time.Duration(interval) * time.Microsecond)
		}
	}
	return nil
}

func main() {
	flag.Parse()
	conf := sarama.NewConfig()
	conf.Producer.Retry.Max = 1
	conf.Producer.RequiredAcks = sarama.WaitForAll
	conf.Producer.Return.Successes = true
	conf.Producer.Timeout = time.Duration(10) * time.Second

	conf.Metadata.Full = true
	conf.Version = sarama.V2_4_0_0
	conf.ClientID = "sasl_scram_client"
	conf.Metadata.Full = true
	conf.Net.SASL.Enable = true
	conf.Net.SASL.User = user
	conf.Net.SASL.Password = password
	conf.Net.SASL.Handshake = true
	conf.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &sasl_sarama.XDGSCRAMClient{HashGeneratorFcn: SHA512} }
	conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512

	splitBrokers := strings.Split(endpoints, ",")
	err := describeTopic(splitBrokers, conf)
	if err == ErrTopicNotExist {
		err = createTopic(splitBrokers, conf)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = produce(splitBrokers, conf)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
