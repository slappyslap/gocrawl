package crawler

import (
	"GoCrawl/internal/log"
	"GoCrawl/internal/redis"
	"GoCrawl/model"
	"context"
	"encoding/json"
	"sync"
	"time"
)

const DB_INSERT_BATCH_SIZE = 1000

type Worker struct {
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewWorker() *Worker {

	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		wg:     &sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,
	}
}

func StartStorageWorker(n int) {
	worker := NewWorker()
	worker.wg.Add(n)

	for i := 0; i < n; i++ {
		go worker.runWorker()
	}

	worker.wg.Wait()
}

func (w *Worker) runWorker() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			results := fetchFromRedis()

			log.Info("Fetched %d results from redis", len(results))

			if len(results) == 0 {
				time.Sleep(time.Second)
			}

			contents := make([]model.Content, 0, DB_INSERT_BATCH_SIZE)

			for _, result := range results {
				var content model.Content
				if err := json.Unmarshal([]byte(result), &content); err != nil {
					log.Error("failed to unmarshal content, err:%s", err.Error())
					continue
				}
				contents = append(contents, content)
			}

			if len(contents) == 0 {
				continue
			}

			err := model.InsertMultipleContents(contents)
			if err != nil {
				continue
			}
		}
	}
}

func fetchFromRedis() []string {
	results := make([]string, 0, DB_INSERT_BATCH_SIZE)

	for i := 0; i < DB_INSERT_BATCH_SIZE; i++ {
		result := redis.LPop(CRAWLER_RESULTS_QUEUE)

		if len(result) == 0 {
			break
		}

		results = append(results, result)
	}

	return results
}
