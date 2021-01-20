package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var stdout = `8.0K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-28
312K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-36
352K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-35
8.0K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-27
8.0K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-33
8.0K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-37
8.0K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-4
9.1G	/var/lib/kafka/data-0/kafka-log1/bench-VM-154-224-centos-15i6-2-0
16K	/var/lib/kafka/data-0/kafka-log1/strimzi.cruisecontrol.partitionmetricsamples-19
1.6M	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-44
7.2G	/var/lib/kafka/data-0/kafka-log1/topic-internal-0
16K	/var/lib/kafka/data-0/kafka-log1/strimzi.cruisecontrol.partitionmetricsamples-6
1.1M	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-19
420K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-0
420K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-7
8.0K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-5
420K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-24
372K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-12
12K	/var/lib/kafka/data-0/kafka-log1/strimzi.cruisecontrol.partitionmetricsamples-20
44K	/var/lib/kafka/data-0/kafka-log1/strimzi.cruisecontrol.partitionmetricsamples-16
8.0K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-15
420K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-9
420K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-3
8.0K	/var/lib/kafka/data-0/kafka-log1/__consumer_offsets-20
17G	/var/lib/kafka/data-0/kafka-log1
17G	/var/lib/kafka/data-0/`

var kubelet_metrics = `kubelet_volume_stats_available_bytes{namespace="sh6-kafka-cluster",persistentvolumeclaim="data-0-kafka-cluster-kafka-3"} 1.05466363904e+11
kubelet_volume_stats_available_bytes{namespace="sh6-kafka-cluster",persistentvolumeclaim="data-1-kafka-cluster-kafka-3"} 1.05465860096e+11
kubelet_volume_stats_available_bytes{namespace="sh6-kafka-cluster",persistentvolumeclaim="data-2-kafka-cluster-kafka-3"} 8.2202021888e+10
kubelet_volume_stats_available_bytes{namespace="sh6-kafka-cluster",persistentvolumeclaim="data-kafka-cluster-zookeeper-1"} 2.0903641088e+10`

const AvailableVolumeLabel = "kubelet_volume_stats_available_bytes"
const PvcLabel = "persistentvolumeclaim"
const NamespaceLabel = "namespace"

type lineArray struct {
	Segements map[string]string
}

type VolumeMetric struct {
	Namespace string
	Pvc       string
	Value     float64
}

func parseTopics(stdout string) []string {
	var topics []string
	lines := strings.Split(stdout, "\n")
	maxSegemnt := 0
	sizeMap := make(map[string]string)
	for _, line := range lines {
		lineSize := strings.Split(line, "\t")
		if len(lineSize) != 2 {
			continue
		}
		segements := strings.Split(lineSize[1], "/")
		if len(segements) > maxSegemnt {
			maxSegemnt = len(segements)
		}
		sizeMap[lineSize[1]] = lineSize[0]
	}
	for line, size := range sizeMap {
		segements := strings.Split(line, "/")
		if len(segements) == maxSegemnt {
			segment := segements[len(segements)-1]
			if strings.HasPrefix(segment, "strimzi.cruisecontrol") {
				continue
			}
			if strings.HasPrefix(segment, "__consumer_offsets") {
				continue
			}
			topics = append(topics, fmt.Sprintf("%s\t%s", size, segment))
		}
	}
	return topics
}

func toVolumeMetrics(b []byte) ([]VolumeMetric, error) {
	var end int
	n := len(b)
	start := 0
	volumeMetrics := make([]VolumeMetric, 0)
	for i := 0; i < n; i++ {
		if b[i] == '\n' || i == n-1 {
			end = i
			if i == n-1 {
				end = i + 1
			}
			if bytes.Contains(b[start:end], []byte(AvailableVolumeLabel)) {
				volumeMetric, err := toVolumeMetric(b[start:end])
				if err != nil {
					return nil, err
				}
				volumeMetrics = append(volumeMetrics, *volumeMetric)
			}
			start = i + 1
		}

	}
	return volumeMetrics, nil
}

func toVolumeMetric(b []byte) (*VolumeMetric, error) {
	n := len(b)
	var start, end int
	var err error
	for i := 0; i < n; i++ {
		if b[i] == '{' {
			start = i + 1
		}
		if b[i] == '}' {
			end = i
			break
		}
	}
	value := string(b[end+2 : n])
	volume := new(VolumeMetric)
	volume.Value, err = strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}
	volume.Value /= 1024 * 1024 * 1024
	labelStr := string(b[start:end])
	labels := strings.Split(labelStr, ",")
	for _, label := range labels {
		labelKv := strings.Split(label, "=")
		if len(labelKv) != 2 {
			continue
		}
		if labelKv[0] == NamespaceLabel {
			volume.Namespace = strings.Trim(labelKv[1], "\"")
		} else if labelKv[0] == PvcLabel {
			volume.Pvc = strings.Trim(labelKv[1], "\"")
		}
	}
	if volume.Pvc == "" || volume.Namespace == "" {
		return nil, errors.New("Unable to parse pvc and namespace for metric")
	}
	return volume, nil
}

func main() {
	// topics := parseTopics(stdout)
	// for _, topic := range topics {
	// 	fmt.Printf("%s\n", topic)
	// }
	volumeMetrics, _ := toVolumeMetrics([]byte(kubelet_metrics))
	// if err != nil {
	// 	fmt.Println(err)
	// }
	fmt.Println(volumeMetrics)
}
