package model

type Response[T any] struct {
	Data    T      `json:"data,omitempty"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}
