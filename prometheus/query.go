package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func QueryKafkaTopic(endpoint, topic string) {
	client, err := api.NewClient(api.Config{
		Address: endpoint, // prometheus url
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := v1api.Query(ctx, "kafka_log_log_size", time.Now())
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	for _, v := range result.(model.Vector) {
		if string(v.Metric["topic"]) == topic { // prometheus label
			fmt.Println(v.Value)
		}
	}
}

func main() {
	QueryKafkaTopic("http://xxx:32000", "topic44")
}
