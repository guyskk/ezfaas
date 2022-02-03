package aliyun

import (
	"encoding/json"

	"github.com/guyskk/ezfaas/internal/common"
)

type AccessConfig struct {
	ALIBABA_CLOUD_ACCOUNT_ID        string
	ALIBABA_CLOUD_ACCESS_KEY_ID     string
	ALIBABA_CLOUD_ACCESS_KEY_SECRET string
}

func LoadAccessConfig() (*AccessConfig, error) {
	data, err := common.ReadUserFile("~/.config/aliyun_fc_deploy.json")
	if err != nil {
		return nil, err
	}
	var config AccessConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
