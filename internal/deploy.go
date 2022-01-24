package internal

import (
	"fmt"
	"log"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/aliyun/fc-go-sdk"
	"github.com/joho/godotenv"
)

type _FunctionConfig struct {
	Region                     string
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
	log.Printf(
		"[INFO] Service=%s Function=%s",
		functionConfig.ServiceName,
		functionConfig.FunctionName,
	)
	log.Printf(
		"[INFO] ContainerImage=%s",
		functionConfig.ContainerImage,
	)
	log.Printf(
		"[INFO] UpdateEnvironmentVariables=%t",
		functionConfig.UpdateEnvironmentVariables,
	)
	endpoint := fmt.Sprintf(
		"%s.%s.fc.aliyuncs.com",
		accessConfig.ALIBABA_CLOUD_ACCOUNT_ID,
		functionConfig.Region,
	)
	log.Printf("[INFO] Deploy Endpoint=%s", endpoint)
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

func _getRegionFromRepository(repository string) string {
	// repository example: registry.cn-zhangjiakou.aliyuncs.com/space/name
	parts := strings.SplitN(repository, ".", 3)
	if len(parts) < 3 {
		log.Fatalf("No region in repository %s", repository)
	}
	return parts[1]
}

type DeployParams struct {
	ServiceName  string
	FunctionName string
	Envfile      string
	Repository   string
	BuildId      string
	Dockerfile   string
	BuildPath    string
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
	if params.BuildId == "" {
		buildResult, err := Build(BuildParams{
			Dockerfile: params.Dockerfile,
			Path:       params.BuildPath,
			Repository: params.Repository,
		})
		if err != nil {
			log.Fatal(err)
		}
		params.BuildId = buildResult.BuildId
	}
	containerImage := fmt.Sprintf("%s:%s", params.Repository, params.BuildId)
	region := _getRegionFromRepository(params.Repository)

	log.Printf("[INFO] Push %s", containerImage)
	err = DockerPush(DockerPushParams{Image: containerImage})
	if err != nil {
		log.Fatal(err)
	}

	functionConfig := _FunctionConfig{
		Region:                     region,
		FunctionName:               params.FunctionName,
		ServiceName:                params.ServiceName,
		ContainerImage:             containerImage,
		UpdateEnvironmentVariables: hasEnv,
		EnvironmentVariables:       env,
	}
	output, err := _updateFunction(accessConfig, &functionConfig)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s \n", output)
}
