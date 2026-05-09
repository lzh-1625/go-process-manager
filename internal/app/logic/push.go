package logic

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/log"
)

type pushLogic struct {
	httpClient *http.Client
}

var PushLogic = &pushLogic{
	httpClient: &http.Client{
		Transport: http.DefaultTransport,
	},
}

func (p *pushLogic) Push(ids []int64, placeholders map[string]string) {
	pl := repository.PushRepository.GetPushConfigByIDs(ids)
	for _, v := range pl {
		if v.Enable {
			var resp *http.Response
			var reader io.Reader = nil
			var url string = p.getReplaceMessage(placeholders, v.Url, true)
			if v.Method == http.MethodPost {
				reader = strings.NewReader(p.getReplaceMessage(placeholders, v.Body, false))
			}
			req, err := http.NewRequest(v.Method, url, reader)
			if err != nil {
				log.Logger.Warnw("push failed", "err", err, "remark", v.Remark)
				continue
			}
			req.Header.Add("content-type", "application/json")
			resp, err = p.httpClient.Do(req)
			if err != nil {
				log.Logger.Warnw("push failed", "err", err, "remark", v.Remark)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
}

func (p *pushLogic) GetPushList() []*model.Push {
	return repository.PushRepository.GetPushList()
}

func (p *pushLogic) GetPushConfigByID(id int) *model.Push {
	return repository.PushRepository.GetPushConfigByID(id)
}

func (p *pushLogic) AddPushConfig(data model.Push) error {
	return repository.PushRepository.AddPushConfig(data)
}

func (p *pushLogic) UpdatePushConfig(data model.Push) error {
	return repository.PushRepository.UpdatePushConfig(data)
}

func (p *pushLogic) DeletePushConfig(id int) error {
	return repository.PushRepository.DeletePushConfig(id)
}

func (p *pushLogic) getReplaceMessage(placeholders map[string]string, message string, urlEncode bool) string {
	kvs := []string{}
	for k, v := range placeholders {
		if urlEncode {
			v = url.QueryEscape(v)
		}
		kvs = append(kvs, k, v)
	}
	return strings.NewReplacer(kvs...).Replace(message)
}
