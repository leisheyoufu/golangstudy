package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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
