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

type PushLogic struct {
	httpClient     *http.Client
	pushRepository *repository.PushRepository
}

func NewPushLogic(pushRepository *repository.PushRepository) *PushLogic {
	return &PushLogic{
		httpClient:     http.DefaultClient,
		pushRepository: pushRepository,
	}
}

func (p *PushLogic) Push(ids []int64, placeholders map[string]string) {
	pl := p.pushRepository.GetPushConfigByIDs(ids)
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

func (p *PushLogic) GetPushList() []*model.Push {
	return p.pushRepository.GetPushList()
}

func (p *PushLogic) GetPushConfigByID(id int) *model.Push {
	return p.pushRepository.GetPushConfigByID(id)
}

func (p *PushLogic) AddPushConfig(data model.Push) error {
	return p.pushRepository.AddPushConfig(data)
}

func (p *PushLogic) UpdatePushConfig(data model.Push) error {
	return p.pushRepository.UpdatePushConfig(data)
}

func (p *PushLogic) DeletePushConfig(id int) error {
	return p.pushRepository.DeletePushConfig(id)
}

func (p *PushLogic) getReplaceMessage(placeholders map[string]string, message string, urlEncode bool) string {
	kvs := []string{}
	for k, v := range placeholders {
		if urlEncode {
			v = url.QueryEscape(v)
		}
		kvs = append(kvs, k, v)
	}
	return strings.NewReplacer(kvs...).Replace(message)
}
