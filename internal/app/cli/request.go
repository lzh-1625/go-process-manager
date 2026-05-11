package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/utils"
)

func Get[T any](uri string) (*T, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1%s/api/process", config.CF.Listen), nil)
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
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://127.0.0.1%s/api/process", config.CF.Listen), bytes.NewReader(jsonBody))
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
	return &result.Data, err
}

func GetJwt() string {
	token, err := utils.GenerateToken(logic.DefaultConsoleAccount, config.CF.SecretKey)
	if err != nil {
		return ""
	}
	return token
}
