package logic

import (
	"io"
	"net/http"
	"strings"

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
	pl := repository.PushRepository.GetPushConfigByIds(ids)
	for _, v := range pl {
		if v.Enable {
			var resp *http.Response
			var reader io.Reader = nil
			var url string = p.getReplaceMessage(placeholders, v.Url)
			if v.Method == http.MethodPost {
				reader = strings.NewReader(p.getReplaceMessage(placeholders, v.Body))
			}
			req, err := http.NewRequest(v.Method, url, reader)
			if err != nil {
				log.Logger.Warnw("推送失败", "err", err, "remark", v.Remark)
				continue
			}
			req.Header.Add("content-type", "application/json")
			resp, err = p.httpClient.Do(req)
			if err != nil {
				log.Logger.Warnw("推送失败", "err", err, "remark", v.Remark)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
}

func (p *pushLogic) getReplaceMessage(placeholders map[string]string, message string) string {
	kvs := []string{}
	for k, v := range placeholders {
		kvs = append(kvs, k, v)
	}
	return strings.NewReplacer(kvs...).Replace(message)
}
