package internal

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/aliyun/fc-go-sdk"
	"github.com/joho/godotenv"
)

type _FunctionConfig struct {
	ServiceName                string
	FunctionName               string
	ContainerImage             string
	UpdateEnvironmentVariables bool
	EnvironmentVariables       map[string]string
}

func _updateFunction(
	accessConfig *AccessConfig,
	functionConfig *_FunctionConfig,
) (*fc.UpdateFunctionOutput, error) {
	endpoint := fmt.Sprintf(
		"%s.cn-zhangjiakou.fc.aliyuncs.com",
		accessConfig.ALIBABA_CLOUD_ACCOUNT_ID,
	)
	client, err := fc.NewClient(
		endpoint,
		"2016-08-15",
		accessConfig.ALIBABA_CLOUD_ACCESS_KEY_ID,
		accessConfig.ALIBABA_CLOUD_ACCESS_KEY_SECRET,
	)
	if err != nil {
		return nil, err
	}
	request := fc.NewUpdateFunctionInput(
		functionConfig.ServiceName,
		functionConfig.FunctionName,
	).WithCustomContainerConfig(
		fc.NewCustomContainerConfig().WithImage(
			functionConfig.ContainerImage,
		),
	)
	if functionConfig.UpdateEnvironmentVariables {
		request = request.WithEnvironmentVariables(
			functionConfig.EnvironmentVariables)
	}
	output, err := client.UpdateFunction(request)
	return output, err
}

type DeployParams struct {
	ServiceName    string
	FunctionName   string
	ContainerImage string
	Envfile        string
}

func DoDeploy(params DeployParams) {
	accessConfig, err := LoadAccessConfig()
	if err != nil {
		log.Fatal(err)
	}
	tomlLoads := toml.Unmarshal
	if tomlLoads == nil {
		fmt.Print("")
	}
	var env map[string]string
	hasEnv := params.Envfile != ""
	if hasEnv {
		envdata, err := ReadUserFile(params.Envfile)
		if err != nil {
			log.Fatal(err)
		}
		env, err = godotenv.Unmarshal(string(envdata))
		if err != nil {
			log.Fatal(err)
		}
	}
	functionConfig := _FunctionConfig{
		FunctionName:               params.FunctionName,
		ServiceName:                params.ServiceName,
		ContainerImage:             params.ContainerImage,
		UpdateEnvironmentVariables: hasEnv,
		EnvironmentVariables:       env,
	}
	output, err := _updateFunction(accessConfig, &functionConfig)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s \n", output)
}
