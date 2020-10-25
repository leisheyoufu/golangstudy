package main

import (
	"crypto/sha512"
	"flag"
	"fmt"
	"hash"
	"log"
	"os"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/leisheyoufu/golangstudy/kafka/sasl_sarama"
	"github.com/xdg/scram"
)

var (
	endpoints string
	user      string
	password  string
	topic     string
	SHA512    scram.HashGeneratorFcn = func() hash.Hash { return sha512.New() }
	logger                           = log.New(os.Stdout, "[Producer] ", log.LstdFlags)
	active                           = true
)

func init() {
	flag.StringVar(&endpoints, "endpoints", "", "Endpoint for kafka, format: 192.168.126.151:9092")
	flag.StringVar(&user, "user", "admin", "kafka user")
	flag.StringVar(&password, "pass", "", "kafka user password")
	flag.StringVar(&topic, "topic", "demo", "kafka topic")
	//sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
}

type KafkaClient struct {
	client           sarama.Client
	groupsPerBroker  map[*sarama.Broker][]string
	topic2Partitions map[string][]int32 // topic -> partition
	topic2Groups     map[string][]string
	offsetManager    *OffsetManager
}

type PartitionOffset struct {
	ParttionOffset map[int32]int64
}

func NewPartitionOffset() *PartitionOffset {
	p := new(PartitionOffset)
	p.ParttionOffset = make(map[int32]int64)
	return p
}

type OffsetManager struct {
	endOffsets   map[string]*PartitionOffset            // topic -> partition offset
	groupOffsets map[string]map[string]*PartitionOffset // topic -> group -> partition offset
}

func NewOffsetManager() *OffsetManager {
	o := new(OffsetManager)
	return o
}

func (om *OffsetManager) RefreshEndOffset(client sarama.Client, topic2Partitions map[string][]int32) {
	om.endOffsets = make(map[string]*PartitionOffset)
	for topic, partitions := range topic2Partitions {
		om.endOffsets[topic] = NewPartitionOffset()
		for _, partition := range partitions {
			offset, err := client.GetOffset(topic, partition, sarama.OffsetNewest)
			if err != nil {
				continue
			}
			om.endOffsets[topic].ParttionOffset[partition] = offset
		}
	}
}

func (om *OffsetManager) RefreshGroupOffset(client sarama.Client, topic2Partitions map[string][]int32, topic2Groups map[string][]string) {
	om.groupOffsets = make(map[string]map[string]*PartitionOffset)
	for topic, groups := range topic2Groups {
		om.groupOffsets[topic] = make(map[string]*PartitionOffset)
		for _, group := range groups {
			manager, _ := sarama.NewOffsetManagerFromClient(group, client)
			defer manager.Close()
			om.groupOffsets[topic][group] = NewPartitionOffset()
			for _, partition := range topic2Partitions[topic] {
				pmanager, _ := manager.ManagePartition(topic, partition)
				offset, _ := pmanager.NextOffset()
				om.groupOffsets[topic][group].ParttionOffset[partition] = offset
			}
		}
	}
}

func (om *OffsetManager) PrintEnd() {
	for topic, partitionOffset := range om.endOffsets {
		for patition, offset := range partitionOffset.ParttionOffset {
			fmt.Printf("Topic %s\t partition %d\t offset %d\n", topic, patition, offset)
		}
	}
}

func (om *OffsetManager) PrintGroup() {
	for topic, groupPartitionMap := range om.groupOffsets {
		for group, partitionOffset := range groupPartitionMap {
			for patition, offset := range partitionOffset.ParttionOffset {
				fmt.Printf("Topic %s\t group %s\t partition %d\t offset %d\n", topic, group, patition, offset)
			}
		}
	}
}

func NewKafkaClient() (*KafkaClient, error) {
	conf := sarama.NewConfig()
	conf.Producer.Retry.Max = 1
	conf.Producer.RequiredAcks = sarama.WaitForAll
	conf.Producer.Return.Successes = true
	conf.Metadata.Full = true
	conf.Version = sarama.V0_10_0_0
	conf.ClientID = "sasl_scram_client"
	conf.Metadata.Full = true
	conf.Net.SASL.Enable = true
	conf.Net.SASL.User = user
	conf.Net.SASL.Password = password
	conf.Net.SASL.Handshake = true
	conf.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &sasl_sarama.XDGSCRAMClient{HashGeneratorFcn: SHA512} }
	conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512

	splitBrokers := strings.Split(endpoints, ",")
	kafkaClient := new(KafkaClient)
	var err error
	kafkaClient.client, err = sarama.NewClient(splitBrokers, conf)
	if err != nil {
		return nil, err
	}
	kafkaClient.groupsPerBroker = make(map[*sarama.Broker][]string)
	kafkaClient.topic2Partitions = make(map[string][]int32)
	kafkaClient.topic2Groups = make(map[string][]string)
	kafkaClient.offsetManager = NewOffsetManager()
	return kafkaClient, nil
}

func (c *KafkaClient) RefreshBrokers() {
	for _, broker := range c.client.Brokers() {
		if ok, _ := broker.Connected(); !ok {
			broker.Open(c.client.Config())
		}
		resp, err := broker.ListGroups(&sarama.ListGroupsRequest{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "ListGroups error : %v\n", err)
			continue
		}
		for group := range resp.Groups {
			c.groupsPerBroker[broker] = append(c.groupsPerBroker[broker], group)
		}
	}
	for k, v := range c.groupsPerBroker {
		fmt.Printf("broker %s -> group %s \n", k.Addr(), v)
	}
}

func (c *KafkaClient) RefreshTopic2Partitions() {
	topics, _ := c.client.Topics()
	//filter topic by topicFilter
	for _, topic := range topics {
		if topic == "__consumer_offsets" {
			continue
		}
		partitions, _ := c.client.Partitions(topic)

		c.topic2Partitions[topic] = partitions
	}
	for k, v := range c.topic2Partitions {
		fmt.Printf("topic: %s -> partitions %v\n", k, v)
	}
}

func (c *KafkaClient) RefreshTopic2Groups() {
	for broker, brokerGroups := range c.groupsPerBroker {
		response, err := broker.DescribeGroups(&sarama.DescribeGroupsRequest{
			Groups: brokerGroups,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "get groupDescribe fail:%v\n", err)
			continue
		}
		topic2Groups := make(map[string]map[string]bool)
		for _, desc := range response.Groups {
			for _, gmd := range desc.Members {
				metadata, err := gmd.GetMemberMetadata()
				if err != nil {
					continue
				}
				for _, topic := range metadata.Topics {
					if _, ok := c.topic2Partitions[topic]; !ok {
						continue
					}
					if _, ok := topic2Groups[topic]; !ok {
						topic2Groups[topic] = make(map[string]bool)
					}
					topic2Groups[topic][desc.GroupId] = true
				}
			}
		}
		for topic, consumerMap := range topic2Groups {
			for group, _ := range consumerMap {
				if _, ok := c.topic2Groups[topic]; !ok {
					c.topic2Groups[topic] = make([]string, 0)
				}
				c.topic2Groups[topic] = append(c.topic2Groups[topic], group)
			}
		}
	}
	for k, v := range c.topic2Groups {
		fmt.Printf("Topic %s: groups %v\n", k, v)
	}
}

func (c *KafkaClient) RefreshOffset() {
	c.offsetManager.RefreshEndOffset(c.client, c.topic2Partitions)
	c.offsetManager.RefreshGroupOffset(c.client, c.topic2Partitions, c.topic2Groups)
	c.offsetManager.PrintEnd()
	c.offsetManager.PrintGroup()
}
func main() {
	flag.Parse()
	kafkaClient, err := NewKafkaClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
	kafkaClient.RefreshBrokers()
	kafkaClient.RefreshTopic2Partitions()
	kafkaClient.RefreshTopic2Groups()
	kafkaClient.RefreshOffset()
}
