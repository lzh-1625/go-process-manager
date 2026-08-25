package victorialogs

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"
)

const (
	insertPath = "/insert/jsonline"
	queryPath  = "/select/logsql/query"
	hitsPath   = "/select/logsql/hits"
)

type victoriaLogsSearch struct {
	client   *http.Client
	url      string
	username string
	password string
}

type victoriaLog struct {
	ID    string `json:"id"`
	Log   string `json:"_msg"`
	Time  string `json:"_time"`
	Name  string `json:"name"`
	Using string `json:"using"`
}

type victoriaLogInsert struct {
	ID    string `json:"id"`
	Log   string `json:"log"`
	Time  string `json:"time"`
	Name  string `json:"name"`
	Using string `json:"using"`
}

func NewVictoriaLogsSearch() search.ILogLogic {
	v := &victoriaLogsSearch{}
	if err := v.init(); err != nil {
		log.Logger.Warnw("Failed to initialize VictoriaLogs client", "err", err)
	}
	return v
}

func (v *victoriaLogsSearch) Reload() error {
	return v.init()
}

func (v *victoriaLogsSearch) init() error {

	v.url = strings.TrimRight("http://xcon.top:9428", "/")

	v.client = &http.Client{
		Transport: &http.Transport{
			IdleConnTimeout: 90 * time.Second,
		},
	}
	return nil
}

func (v *victoriaLogsSearch) Insert(logs ...model.ProcessLog) {
	if len(logs) == 0 || v.client == nil {
		return
	}

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
			log.Logger.Errorw("VictoriaLogs encode failed", "err", err)
			return
		}
	}

	endpoint := v.endpoint(insertPath)
	parsed, err := url.Parse(endpoint)
	if err != nil {
		log.Logger.Errorw("VictoriaLogs insert URL is invalid", "err", err)
		return
	}
	query := parsed.Query()
	query.Set("_time_field", "time")
	query.Set("_msg_field", "log")
	parsed.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, parsed.String(), &body)
	if err != nil {
		log.Logger.Errorw("VictoriaLogs insert request failed", "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/stream+json")
	v.setAuth(req)
	if err := v.do(req); err != nil {
		log.Logger.Errorw("VictoriaLogs insert failed", "err", err)
	}
}

func (v *victoriaLogsSearch) Search(req model.GetLogReq) (result model.LogResp) {
	if v.client == nil {
		log.Logger.Warn("VictoriaLogs client is not initialized")
		return result
	}

	query := buildQuery(req)
	data, err := v.query(query, req)
	if err != nil {
		log.Logger.Warnw("VictoriaLogs search failed", "err", err)
		return result
	}
	if req.Match.HighLight {
		highlight(data, search.QueryStringAnalysis(req.Match.Log))
	}
	result.Data = data

	total, err := v.hits(buildFilters(req), req)
	if err != nil {
		log.Logger.Warnw("VictoriaLogs hits query failed", "err", err)
		return result
	}
	result.Total = total
	return result
}

func (v *victoriaLogsSearch) query(query string, req model.GetLogReq) ([]*model.ProcessLog, error) {
	values := url.Values{"query": {query}}
	setTimeRange(values, req)
	response, err := v.postForm(queryPath, values)
	if err != nil {
		return nil, err
	}
	defer response.Close()

	logs := make([]*model.ProcessLog, 0)
	scanner := bufio.NewScanner(response)
	buffer := make([]byte, 64*1024)
	scanner.Buffer(buffer, 4*1024*1024)
	for scanner.Scan() {
		var entry victoriaLog
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return nil, fmt.Errorf("decode VictoriaLogs query response: %w", err)
		}
		id, err := strconv.ParseInt(entry.ID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse VictoriaLogs log id: %w", err)
		}
		entryTime, err := time.Parse(time.RFC3339Nano, entry.Time)
		if err != nil {
			return nil, fmt.Errorf("parse VictoriaLogs log time: %w", err)
		}
		logs = append(logs, &model.ProcessLog{
			ID:    id,
			Log:   entry.Log,
			Time:  entryTime.UnixMilli(),
			Name:  entry.Name,
			Using: entry.Using,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read VictoriaLogs query response: %w", err)
	}
	return logs, nil
}

func (v *victoriaLogsSearch) hits(query string, req model.GetLogReq) (int64, error) {
	values := url.Values{"query": {query}}
	setTimeRange(values, req)
	values.Set("step", "1h")
	response, err := v.postForm(hitsPath, values)
	if err != nil {
		return 0, err
	}
	defer response.Close()

	var body struct {
		Hits []struct {
			Total int64 `json:"total"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(response).Decode(&body); err != nil {
		return 0, fmt.Errorf("decode VictoriaLogs hits response: %w", err)
	}
	var total int64
	for _, hit := range body.Hits {
		total += hit.Total
	}
	return total, nil
}

func (v *victoriaLogsSearch) postForm(path string, values url.Values) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(
		context.TODO(),
		http.MethodPost,
		v.endpoint(path),
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	v.setAuth(req)

	response, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		defer response.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4*1024))
		return nil, fmt.Errorf("VictoriaLogs returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	return response.Body, nil
}

func (v *victoriaLogsSearch) do(req *http.Request) error {
	response, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4*1024))
		return fmt.Errorf("VictoriaLogs returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (v *victoriaLogsSearch) endpoint(path string) string {
	return v.url + path
}

func (v *victoriaLogsSearch) setAuth(req *http.Request) {
	if v.username != "" || v.password != "" {
		req.SetBasicAuth(v.username, v.password)
	}
}

func buildQuery(req model.GetLogReq) string {
	query := buildFilters(req)
	sort := "sort by (_time, id)"
	if req.Sort == "desc" {
		sort = "sort by (_time desc, id desc)"
	}
	return fmt.Sprintf("%s | %s offset %d limit %d | fields _msg, _time, id, name, using", query, sort, req.Page.From, req.Page.Size)
}

func buildFilters(req model.GetLogReq) string {
	filters := []string{"*"}
	for _, item := range search.QueryStringAnalysis(req.Match.Log) {
		filter := ""
		switch item.Cond {
		case search.Match:
			filter = "i(" + quote(item.Content) + ")"
		case search.NotMatch:
			filter = "NOT i(" + quote(item.Content) + ")"
		case search.WildCard:
			filter = "~" + quote("(?i)"+regexp.QuoteMeta(item.Content))
		case search.NotWildCard:
			filter = "NOT ~" + quote("(?i)"+regexp.QuoteMeta(item.Content))
		}
		if filter != "" {
			filters = append(filters, filter)
		}
	}
	if req.Match.Name != "" {
		filters = append(filters, "name:="+quote(req.Match.Name))
	}
	if req.Match.Using != "" {
		filters = append(filters, "using:="+quote(req.Match.Using))
	}
	if req.CursorID != 0 {
		operator := ">"
		if req.Sort == "desc" {
			operator = "<"
		}
		filters = append(filters, "id:"+operator+strconv.FormatInt(req.CursorID, 10))
	}
	if len(req.FilterName) != 0 {
		values := make([]string, 0, len(req.FilterName))
		for _, name := range req.FilterName {
			values = append(values, quote(name))
		}
		filters = append(filters, "name:in("+strings.Join(values, ", ")+")")
	}
	return strings.Join(filters, " ")
}

func quote(value string) string {
	return strconv.Quote(value)
}

func setTimeRange(values url.Values, req model.GetLogReq) {
	if req.TimeRange.StartTime != 0 {
		values.Set("start", time.UnixMilli(req.TimeRange.StartTime).UTC().Format(time.RFC3339Nano))
	}
	if req.TimeRange.EndTime != 0 {
		values.Set("end", time.UnixMilli(req.TimeRange.EndTime).UTC().Format(time.RFC3339Nano))
	}
}

func highlight(logs []*model.ProcessLog, query []search.Query) {
	for _, item := range query {
		if item.Cond == search.NotMatch || item.Cond == search.NotWildCard {
			continue
		}
		for _, log := range logs {
			log.Log = utils.StringReplaceHighLight(log.Log, item.Content)
		}
	}
}
