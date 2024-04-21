package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/siddontang/go-log/log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	AvgCpu float64
	MaxCpu float64
	AvgMem float64
	MaxMem float64
}

type NodeInfo struct {
	Name         string
	RequestedCpu int64
	CapacityCpu  int64
	RequestedMem int64
	CapacityMem  int64
	UsageCpu     int64
	UsageMem     int64
}

type PodInfo struct {
	Name       string
	Task       string
	RequestCpu int64 // MilliValue
	LimitCpu   int64 // MilliValue
	RequestMem int64
	LimitMem   int64
	Status     string
	Metric     *Metrics
	NodeName   string
	Node       *NodeInfo
}

type Cred struct {
	endpoint string
	user     string
	password string
	region   string
}

func GetAvgMatrixValue(result model.Value) float64 {
	matrix := result.(model.Matrix)
	var sum float64
	n := 0
	for _, sample := range matrix {
		for _, point := range sample.Values {
			sum += float64(point.Value)
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

func GetMaxMatrixValue(result model.Value) float64 {
	matrix := result.(model.Matrix)
	var ret float64
	for _, sample := range matrix {
		for _, point := range sample.Values {
			if float64(point.Value) > ret {
				ret = float64(point.Value)
			}
		}
	}
	return ret
}

func GetVectorValue(result model.Value) float64 {
	vector := result.(model.Vector)
	for _, sample := range vector {
		return float64(sample.Value)
	}
	return 0
}

func GetMetric(ctx context.Context, cred *Cred, pod *PodInfo) (*Metrics, error) {
	username := cred.user
	password := cred.password

	// 创建一个 Prometheus 客户端
	client, err := api.NewClient(api.Config{
		Address:      cred.endpoint,
		RoundTripper: config.NewBasicAuthRoundTripper(username, config.Secret(password), "", "", api.DefaultRoundTripper),
	})
	if err != nil {
		log.Errorf("prometheus pod %s region %s err: %v", pod.Name, cred.region, err)
		return nil, err
	}

	// 创建一个 Prometheus API 客户端
	promClient := v1.NewAPI(client)
	metrics := new(Metrics)
	// avg cpu
	// 定义要查询的指标名称和标签
	var query string
	ctx, _ = context.WithTimeout(ctx, 10*time.Second)
	query = fmt.Sprintf("avg_over_time(rate(process_cpu_seconds_total{pod=\"%s\"}[1m])[24h:])", pod.Name)
	result, warnings, err := promClient.QueryRange(ctx, query, v1.Range{
		Start: time.Now().Add(-time.Hour * 24),
		End:   time.Now(),
		Step:  time.Minute}, v1.WithTimeout(10*time.Second))
	if err != nil {
		log.Errorf("failed to get pod %s, endpoint %s, err: %v", pod.Name, cred.endpoint, err)
		return nil, err
	}
	if len(warnings) != 0 {
		for _, warning := range warnings {
			log.Warnf("get warning %s while getting pod %s, endpoint %s", warning, pod.Name, cred.endpoint)
		}
	}
	metrics.AvgCpu = GetAvgMatrixValue(result)
	// max cpu
	query = fmt.Sprintf("max_over_time(rate(process_cpu_seconds_total{pod='%s'}[1m])[24h:])", pod.Name)
	result, warnings, err = promClient.QueryRange(context.Background(), query, v1.Range{
		Start: time.Now().Add(-time.Hour * 24),
		End:   time.Now(),
		Step:  time.Minute}, v1.WithTimeout(10*time.Second))
	if err != nil {
		log.Errorf("failed to get pod %s, endpoint %s", pod.Name, cred.endpoint)
		return nil, err
	}
	if len(warnings) != 0 {
		for _, warning := range warnings {
			log.Warnf("get warning %s while getting pod %s, endpoint %s", warning, pod.Name, cred.endpoint)
		}
	}
	metrics.MaxCpu = GetMaxMatrixValue(result)
	// avg mem
	query = fmt.Sprintf("avg_over_time(process_resident_memory_bytes{pod=\"%s\"}[24h:]) / 1024/1024", pod.Name)
	result, warnings, err = promClient.QueryRange(context.Background(), query, v1.Range{
		Start: time.Now().Add(-time.Hour * 24),
		End:   time.Now(),
		Step:  time.Minute}, v1.WithTimeout(10*time.Second))
	if err != nil {
		log.Errorf("failed to get pod %s, endpoint %s", pod.Name, cred.endpoint)
		return nil, err
	}
	if len(warnings) != 0 {
		for _, warning := range warnings {
			log.Warnf("get warning %s while getting pod %s, endpoint %s", warning, pod.Name, cred.endpoint)
		}
	}
	metrics.AvgMem = GetAvgMatrixValue(result)

	// max mem
	query = fmt.Sprintf("max_over_time(process_resident_memory_bytes{pod=\"%s\"}[24h:]) / 1024/1024", pod.Name)
	result, warnings, err = promClient.QueryRange(context.Background(), query, v1.Range{
		Start: time.Now().Add(-time.Hour * 24),
		End:   time.Now(),
		Step:  time.Minute}, v1.WithTimeout(10*time.Second))
	if err != nil {
		log.Errorf("failed to get pod %s, endpoint %s", pod.Name, cred.endpoint)
		return nil, err
	}
	if len(warnings) != 0 {
		for _, warning := range warnings {
			log.Warnf("get warning %s while getting pod %s, endpoint %s", warning, pod.Name, cred.endpoint)
		}
	}
	metrics.MaxMem = GetMaxMatrixValue(result)
	return metrics, nil
}

func main() {
	// 定义一个Histogram类型的指标
	histogram := promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "histogram_showcase_metric",
		Buckets: []float64{5.0, 10.0, 20.0, 50.0, 100.0}, // 根据场景需求配置bucket的范围
	})

	go func() {
		for {
			// 这里搜集一些0-100之间的随机数
			// 实际应用中，这里可以搜集系统耗时等指标
			histogram.Observe(rand.Float64() * 100.0)
			time.Sleep(1 * time.Second)
		}
	}()
	// 指标上报的路径，可以通过该路径获取实时的监控数据
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
