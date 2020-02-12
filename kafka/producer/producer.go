package main

import (
	"flag"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"time"
)

var (
	endpoint string
)

func syncProducer(endpoints []string) {
	config := sarama.NewConfig()
	// 等待服务器所有副本都保存成功后的响应
	config.Producer.RequiredAcks = sarama.WaitForAll
	// 随机的分区类型：返回一个分区器，该分区器每次选择一个随机分区
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	// 是否等待成功和失败后的响应
	config.Producer.Return.Successes = true
	// 使用给定代理地址和配置创建一个同步生产者
	producer, err := sarama.NewSyncProducer([]string{endpoint}, config)
	if err != nil {
		panic(err)
	}

	defer producer.Close()

	topic := "demo"
	srcValue := "sync: this is a message. index=%d"
	for i := 0; i < 10; i++ {
		value := fmt.Sprintf(srcValue, i)
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(value),
		}
		part, offset, err := producer.SendMessage(msg)
		if err != nil {
			log.Printf("send message(%s) err=%s \n", value, err)
		} else {
			log.Printf("send message successfully, part=%d, offset=%d\n", part, offset)
		}
		time.Sleep(1 * time.Second)
	}

}

func main() {
	flag.Parse()
	syncProducer([]string{endpoint})
}

func init() {
	flag.StringVar(&endpoint, "endpoint", "", "Endpoint for kafka, format: 192.168.126.151:9092")
}
