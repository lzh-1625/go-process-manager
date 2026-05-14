package model

type SystemConfigurationVo struct {
	Key      string `json:"key"`
	Value    any    `json:"value"`
	Default  string `json:"default"`
	Describe string `json:"describe"`
}
