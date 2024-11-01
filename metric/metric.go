package metric

import (
	"GoCrawl/crawler"
	"GoCrawl/internal/log"
	"GoCrawl/internal/redis"
	"GoCrawl/model"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"sync"
	"time"
)

type Metrics struct {
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

var (
	totalUrlCrawled = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_url_crawled",
		Help: "Total number of urls crawled",
	})

	differentHosts = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "different_hosts",
		Help: "Total number of different hosts crawled",
	})

	currentQueueSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "current_queue_size",
		Help: "Current size of the queue",
	})

	storageQueueSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "storage_queue_size",
		Help: "Current size of the storage queue",
	})
)

func init() {
	prometheus.MustRegister(totalUrlCrawled)
	prometheus.MustRegister(differentHosts)
	prometheus.MustRegister(currentQueueSize)
	prometheus.MustRegister(storageQueueSize)
}

func ExposeMetrics() {

	go func() {
		for {
			updateMetrics()
			time.Sleep(time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":2112", nil)
	if err != nil {
		log.Error("failed to expose metrics, err: %s", err.Error())
		return
	}
}

func updateMetrics() {
	updateTotalUrlCrawled()
	updateDifferentHosts()
	updateCurrentQueueSize()
	updateStorageQueueSize()
}

func updateTotalUrlCrawled() {
	totalUrlCrawledCount, err := model.GetContentCount()

	if err != nil {
		log.Error("failed to get total url crawled count, err: %s", err.Error())
		return
	}

	totalUrlCrawled.Set(float64(totalUrlCrawledCount))
}

func updateDifferentHosts() {
	count := redis.HLen(crawler.CRAWLER_HOST_COUNTER)
	differentHosts.Set(float64(count))
}

func updateCurrentQueueSize() {
	currentQueueSizeCount := redis.GetQueueSize(crawler.CRAWLER_TASK_QUEUE)
	currentQueueSize.Set(float64(currentQueueSizeCount))
}

func updateStorageQueueSize() {
	storageQueueSizeCount := redis.GetQueueSize(crawler.CRAWLER_RESULTS_QUEUE)
	storageQueueSize.Set(float64(storageQueueSizeCount))
}
