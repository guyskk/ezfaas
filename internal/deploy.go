package internal

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/guyskk/ezfaas/internal/aliyun"
	"github.com/guyskk/ezfaas/internal/common"
	"github.com/guyskk/ezfaas/internal/tencent"
	"github.com/joho/godotenv"
)

type BaseDeployParams struct {
	FunctionName string
	Envfile      string
	Repository   string
	BuildId      string
	Dockerfile   string
	BuildPath    string
	Yes          bool
}

type AliyunDeployParams struct {
	BaseDeployParams
	ServiceName string
}

type TencentDeployParams struct {
	BaseDeployParams
	Region string
}

func _readEnvfile(envfile string) *map[string]string {
	tomlLoads := toml.Unmarshal
	if tomlLoads == nil {
		fmt.Print("")
	}
	var env *map[string]string = nil
	if envfile != "" {
		envdata, err := common.ReadUserFile(envfile)
		if err != nil {
			log.Fatal(err)
		}
		_env, err := godotenv.Unmarshal(string(envdata))
		if err != nil {
			log.Fatal(err)
		}
		env = &_env
	}
	return env
}

func _prepareImage(params BaseDeployParams) string {
	buildId := params.BuildId
	if buildId == "" {
		buildResult, err := Build(BuildParams{
			Dockerfile: params.Dockerfile,
			Path:       params.BuildPath,
			Repository: params.Repository,
		})
		if err != nil {
			log.Fatal(err)
		}
		buildId = buildResult.BuildId
	}
	containerImage := fmt.Sprintf("%s:%s", params.Repository, buildId)
	log.Printf("[INFO] Push %s", containerImage)
	err := common.DockerPush(common.DockerPushParams{Image: containerImage})
	if err != nil {
		log.Fatal(err)
	}
	return buildId
}

func DoDeployAliyun(params AliyunDeployParams) {
	env := _readEnvfile(params.Envfile)
	buildId := _prepareImage(params.BaseDeployParams)
	output, err := aliyun.DoDeploy(aliyun.DeployParams{
		ServiceName:          params.ServiceName,
		FunctionName:         params.FunctionName,
		Repository:           params.Repository,
		Yes:                  params.Yes,
		BuildId:              buildId,
		EnvironmentVariables: env,
	})
	if err != nil {
		log.Fatal(err)
	}
	outputBytes, _ := json.MarshalIndent(output, "", "    ")
	log.Printf("%s\n", string(outputBytes))
}

func DoDeployTencent(params TencentDeployParams) {
	env := _readEnvfile(params.Envfile)
	buildId := _prepareImage(params.BaseDeployParams)
	output, err := tencent.DoDeploy(tencent.DeployParams{
		Region:               params.Region,
		FunctionName:         params.FunctionName,
		Repository:           params.Repository,
		Yes:                  params.Yes,
		BuildId:              buildId,
		EnvironmentVariables: env,
	})
	if err != nil {
		log.Fatal(err)
	}
	outputBytes, _ := json.MarshalIndent(output, "", "    ")
	log.Printf("%s\n", string(outputBytes))
}
