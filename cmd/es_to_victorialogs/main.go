package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/olivere/elastic/v7"
)

const (
	defaultScrollBatchSize = 5000
	insertPath             = "/insert/jsonline"
)

type migrationConfig struct {
	esURL      string
	esIndex    string
	esUsername string
	esPassword string
	vlURL      string
	vlUsername string
	vlPassword string
	batchSize  int
}

type victoriaLogInsert struct {
	ID    string `json:"id"`
	Log   string `json:"log"`
	Time  string `json:"time"`
	Name  string `json:"name"`
	Using string `json:"using"`
}

func main() {
	cfg := parseFlags()
	if err := cfg.validate(); err != nil {
		log.Fatal(err)
	}
	if err := migrate(context.Background(), cfg); err != nil {
		log.Fatal(err)
	}
}

func parseFlags() migrationConfig {
	var cfg migrationConfig
	flag.StringVar(&cfg.esURL, "es-url", "http://xcon.top:9200", "Elasticsearch URL (or GPM_ES_URL)")
	flag.StringVar(&cfg.esIndex, "es-index", "server_log_v2", "Elasticsearch index (or GPM_ES_INDEX)")
	flag.StringVar(&cfg.esUsername, "es-username", "elastic", "Elasticsearch username (or GPM_ES_USERNAME)")
	flag.StringVar(&cfg.esPassword, "es-password", "1625167628@xcon", "Elasticsearch password (or GPM_ES_PASSWORD)")
	flag.StringVar(&cfg.vlURL, "victorialogs-url", "http://xcon.top:9428", "VictoriaLogs URL (or GPM_VICTORIALOGS_URL)")
	flag.StringVar(&cfg.vlUsername, "victorialogs-username", "", "VictoriaLogs username (or GPM_VICTORIALOGS_USERNAME)")
	flag.StringVar(&cfg.vlPassword, "victorialogs-password", "", "VictoriaLogs password (or GPM_VICTORIALOGS_PASSWORD)")
	flag.IntVar(&cfg.batchSize, "batch-size", defaultScrollBatchSize, "ES scroll and VictoriaLogs insert batch size")
	flag.Parse()
	return cfg
}

func (c migrationConfig) validate() error {
	for _, target := range []struct {
		name  string
		value string
	}{
		{"es-url", c.esURL},
		{"victorialogs-url", c.vlURL},
	} {
		parsed, err := url.ParseRequestURI(target.value)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("%s must be an absolute URL", target.name)
		}
	}
	if c.esIndex == "" {
		return errors.New("es-index is required")
	}
	if c.batchSize <= 0 {
		return errors.New("batch-size must be positive")
	}
	return nil
}

func migrate(ctx context.Context, cfg migrationConfig) error {
	esClient, err := newESClient(cfg)
	if err != nil {
		return fmt.Errorf("connect Elasticsearch: %w", err)
	}
	vlClient := &http.Client{
		Timeout: 2 * time.Minute,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 1,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	total, err := esClient.Count(cfg.esIndex).Do(ctx)
	if err != nil {
		log.Printf("get Elasticsearch document count failed: %v", err)
	} else {
		fmt.Printf("index %s has %d documents\n", cfg.esIndex, total)
	}

	start := time.Now()
	var migrated, skipped int64
	scroll := esClient.Scroll(cfg.esIndex).Size(cfg.batchSize).Sort("_doc", true)
	defer func() { _ = scroll.Clear(ctx) }()

	for {
		response, err := scroll.Do(ctx)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("scroll Elasticsearch documents: %w", err)
		}
		if len(response.Hits.Hits) == 0 {
			break
		}

		logs := make([]model.ProcessLog, 0, len(response.Hits.Hits))
		for _, hit := range response.Hits.Hits {
			item, include, err := decodeProcessLog(hit.Id, hit.Source)
			if err != nil {
				skipped++
				log.Printf("skip Elasticsearch document %s: %v", hit.Id, err)
				continue
			}
			if !include {
				skipped++
				log.Printf("skip Elasticsearch document %s: missing or invalid id", hit.Id)
				continue
			}
			logs = append(logs, item)
		}

		if len(logs) > 0 {
			if err := insertVictoriaLogs(ctx, vlClient, cfg, logs); err != nil {
				return fmt.Errorf("insert VictoriaLogs batch starting at %d: %w", migrated, err)
			}
			migrated += int64(len(logs))
		}
		printProgress(migrated, total, start)
	}

	fmt.Printf("\nmigration complete: %d documents written, %d documents skipped in %v\n", migrated, skipped, time.Since(start).Round(time.Second))
	return nil
}

func decodeProcessLog(documentID string, source []byte) (model.ProcessLog, bool, error) {
	var item model.ProcessLog
	if err := json.Unmarshal(source, &item); err != nil {
		return model.ProcessLog{}, false, fmt.Errorf("decode document %s: %w", documentID, err)
	}
	return item, item.ID > 0, nil
}

func newESClient(cfg migrationConfig) (*elastic.Client, error) {
	options := []elastic.ClientOptionFunc{
		elastic.SetURL(cfg.esURL),
		elastic.SetSniff(false),
		elastic.SetHttpClient(&http.Client{Transport: &http.Transport{
			MaxIdleConnsPerHost: 1,
			IdleConnTimeout:     90 * time.Second,
		}}),
	}
	if cfg.esUsername != "" || cfg.esPassword != "" {
		options = append(options, elastic.SetBasicAuth(cfg.esUsername, cfg.esPassword))
	}
	return elastic.NewClient(options...)
}

func insertVictoriaLogs(ctx context.Context, client *http.Client, cfg migrationConfig, logs []model.ProcessLog) error {
	req, err := newVictoriaLogsRequest(ctx, cfg.vlURL, logs)
	if err != nil {
		return err
	}
	if cfg.vlUsername != "" || cfg.vlPassword != "" {
		req.SetBasicAuth(cfg.vlUsername, cfg.vlPassword)
	}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4*1024))
		return fmt.Errorf("VictoriaLogs returned %s: %s", response.Status, strings.TrimSpace(string(message)))
	}
	return nil
}

func newVictoriaLogsRequest(ctx context.Context, victoriaLogsURL string, logs []model.ProcessLog) (*http.Request, error) {
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	for _, item := range logs {
		entry := victoriaLogInsert{
			ID:    strconv.FormatInt(item.ID, 10),
			Log:   item.Log,
			Time:  time.UnixMilli(item.Time).UTC().Format(time.RFC3339Nano),
			Name:  item.Name,
			Using: item.Using,
		}
		if err := encoder.Encode(entry); err != nil {
			return nil, fmt.Errorf("encode log %d: %w", item.ID, err)
		}
	}

	endpoint, err := url.Parse(strings.TrimRight(victoriaLogsURL, "/") + insertPath)
	if err != nil {
		return nil, fmt.Errorf("parse VictoriaLogs URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("_time_field", "time")
	query.Set("_msg_field", "log")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), &body)
	if err != nil {
		return nil, fmt.Errorf("create VictoriaLogs insert request: %w", err)
	}
	req.Header.Set("Content-Type", "application/stream+json")
	return req, nil
}

func printProgress(migrated, total int64, start time.Time) {
	elapsed := time.Since(start).Seconds()
	rate := float64(migrated) / elapsed
	pct := ""
	if total > 0 {
		pct = fmt.Sprintf(" (%.1f%%)", float64(migrated)/float64(total)*100)
	}
	fmt.Printf("\rmigrated: %d/%d%s rate: %.0f docs/s", migrated, total, pct, rate)
}
