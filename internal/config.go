package internal

import (
	"encoding/json"
	"os"
	gofilepath "path/filepath"

	"github.com/mitchellh/go-homedir"
)

type AccessConfig struct {
	ALIBABA_CLOUD_ACCOUNT_ID        string
	ALIBABA_CLOUD_ACCESS_KEY_ID     string
	ALIBABA_CLOUD_ACCESS_KEY_SECRET string
}

func ReadUserFile(filepath string) ([]byte, error) {
	filepath, err := homedir.Expand(filepath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(gofilepath.ToSlash(filepath))
}

func LoadAccessConfig() (*AccessConfig, error) {
	data, err := ReadUserFile("~/.config/aliyun_fc_deploy.json")
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
