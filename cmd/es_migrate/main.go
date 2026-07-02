package main

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/olivere/elastic/v7"
)

const (
	scrollBatchSize = 5000
	bulkBatchSize   = 2000
	workerCount     = 8
)

type updateTask struct {
	docID string
	newID int64
}

func main() {
	esURL := config.CF.EsUrl
	esUser := config.CF.EsUsername
	esPass := config.CF.EsPassword
	esIndex := config.CF.EsIndex

	if cmp.Or(esURL, esUser, esPass, esIndex) == "" {
		log.Fatalf("es url, user, pass, index is not set")
	}

	opts := []elastic.ClientOptionFunc{
		elastic.SetURL(esURL),
		elastic.SetSniff(false),
		elastic.SetHttpClient(&http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: workerCount * 2,
				IdleConnTimeout:     90 * time.Second,
			},
		}),
	}
	if esUser != "" {
		opts = append(opts, elastic.SetBasicAuth(esUser, esPass))
	}

	client, err := elastic.NewClient(opts...)
	if err != nil {
		log.Fatalf("connect es failed: %v", err)
	}

	ctx := context.Background()

	// get total, for progress display
	total, err := client.Count(esIndex).Do(ctx)
	if err != nil {
		log.Printf("get total failed: %v", err)
	} else {
		fmt.Printf("index %s has %d documents\n", esIndex, total)
	}

	// batchCh buffer size = workerCount*2, enough buffer for scroll and worker
	batchCh := make(chan []updateTask, workerCount*2)

	var (
		updatedCount int64
		wg           sync.WaitGroup
	)

	startTime := time.Now()

	// start worker pool, concurrent bulk update
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tasks := range batchCh {
				bulk := client.Bulk()
				for _, t := range tasks {
					bulk = bulk.Add(
						elastic.NewBulkUpdateRequest().
							Index(esIndex).
							Id(t.docID).
							Doc(map[string]any{"id": t.newID}),
					)
				}
				res, bulkErr := bulk.Do(ctx)
				if bulkErr != nil {
					log.Printf("bulk update failed: %v", bulkErr)
					continue
				}
				if res.Errors {
					for _, item := range res.Failed() {
						log.Printf("document update failed: id=%s err=%v", item.Id, item.Error)
					}
				}
				n := atomic.AddInt64(&updatedCount, int64(len(tasks)))
				printProgress(n, total, startTime)
			}
		}()
	}

	// main goroutine: scroll read, assign id sequentially, dispatch to worker
	scroll := client.Scroll(esIndex).
		Size(scrollBatchSize).
		TrackTotalHits(true)

	var (
		idCounter int64
		pending   []updateTask
	)

	for {
		res, scrollErr := scroll.Do(ctx)
		if scrollErr != nil {
			if scrollErr == io.EOF {
				break
			}
			log.Fatalf("scroll failed: %v", scrollErr)
		}
		if len(res.Hits.Hits) == 0 {
			break
		}

		for _, h := range res.Hits.Hits {
			idCounter++
			pending = append(pending, updateTask{docID: h.Id, newID: idCounter})

			if len(pending) >= bulkBatchSize {
				batchCh <- pending
				pending = make([]updateTask, 0, bulkBatchSize)
			}
		}
	}

	// remaining data less than one batch
	if len(pending) > 0 {
		batchCh <- pending
	}

	close(batchCh)
	_ = scroll.Clear(ctx)

	wg.Wait()

	elapsed := time.Since(startTime).Round(time.Second)
	fmt.Printf("\ndone! updated %d documents, elapsed %v\n", idCounter, elapsed)
}

func printProgress(updated, total int64, start time.Time) {
	elapsed := time.Since(start).Seconds()
	rate := float64(updated) / elapsed
	var eta string
	if total > 0 && rate > 0 {
		remaining := float64(total-updated) / rate
		eta = fmt.Sprintf("remaining %v", time.Duration(remaining*float64(time.Second)).Round(time.Second))
	}
	pct := ""
	if total > 0 {
		pct = fmt.Sprintf("(%.1f%%)", float64(updated)/float64(total)*100)
	}
	fmt.Printf("\rupdated: %d/%d %s  rate: %.0f doc/s%s   ", updated, total, pct, rate, eta)
}
