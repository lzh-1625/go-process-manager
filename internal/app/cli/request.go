package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/utils"
)

func Get[T any](uri string, query map[string]string) (*T, error) {
	q := url.Values{}
	for k, v := range query {
		q.Add(k, v)
	}
	u := url.URL{
		Scheme:   "http",
		Host:     config.CF.Listen,
		Path:     uri,
		RawQuery: q.Encode(),
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	return Do[T](req)
}

func Put[T any](uri string, body any) (*T, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	u := url.URL{
		Scheme: "http",
		Host:   config.CF.Listen,
		Path:   uri,
	}
	req, err := http.NewRequest(http.MethodPut, u.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	return Do[T](req)
}

func Delete[T any](uri string, query map[string]string) (*T, error) {
	q := url.Values{}
	for k, v := range query {
		q.Add(k, v)
	}
	u := url.URL{
		Scheme:   "http",
		Host:     config.CF.Listen,
		Path:     uri,
		RawQuery: q.Encode(),
	}
	req, err := http.NewRequest(http.MethodDelete, u.String(), nil)
	if err != nil {
		return nil, err
	}
	return Do[T](req)
}

func Post[T any](uri string, body any) (*T, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	u := url.URL{
		Scheme: "http",
		Host:   config.CF.Listen,
		Path:   uri,
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	return Do[T](req)
}

func Do[T any](req *http.Request) (*T, error) {
	req.Header.Set("Authorization", "bearer "+GetJwt())
	req.Header.Set("Content-Type", "application/json")
	var result model.Response[T]
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, errors.New(result.Message)
	}
	return &result.Data, err
}

func GetJwt() string {
	token, err := utils.GenerateToken(logic.DefaultConsoleAccount, config.CF.SecretKey, time.Now().Add(time.Hour))
	if err != nil {
		return ""
	}
	return token
}
