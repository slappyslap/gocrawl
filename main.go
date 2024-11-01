package main

import (
	"GoCrawl/crawler"
	"GoCrawl/internal/log"
	"GoCrawl/internal/redis"
	"GoCrawl/metric"
	"flag"
	"os"
	"os/signal"
	"syscall"
)

var workerMap map[string]func(count int) = map[string]func(count int){
	"storageWorker": crawler.StartStorageWorker,
	"crawler":       crawler.Scrape,
}

func main() {
	shutdown := make(chan int)
	signal.Notify(crawler.Sigchan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Handle shutdown signal from the OS like (Ctrl + C)
	go func() {
		<-crawler.Sigchan
		shutdown <- 1
	}()

	var worker string
	var count int

	flag.StringVar(&worker, "w", "crawler", "Choose the worker to run")
	flag.IntVar(&count, "c", 2, "Number of workers to run")

	var startUrl string
	flag.StringVar(&startUrl, "url", "", "Choose the start url")

	flag.Parse()

	if startUrl != "" {
		err := redis.LPush(crawler.CRAWLER_TASK_QUEUE, startUrl)
		if err != nil {
			log.Error("failed to push start url %s to queue %s, err:%s", startUrl, crawler.CRAWLER_TASK_QUEUE, err.Error())
			return
		}
	}

	if worker == "crawler" {
		log.Info("Starting metrics server")
		go metric.ExposeMetrics()
	}

	f, ok := workerMap[worker]
	if !ok {
		panic("worker not found")
	}
	go f(count)

	<-shutdown
}
