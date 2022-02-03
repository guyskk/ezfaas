package aliyun

import (
	"fmt"
	"log"
	"strings"

	"github.com/aliyun/fc-go-sdk"
	"github.com/guyskk/ezfaas/internal/common"
)

type _FunctionConfig struct {
	Region                     string
	ServiceName                string
	FunctionName               string
	ContainerImage             string
	UpdateEnvironmentVariables bool
	EnvironmentVariables       map[string]string
	Yes                        bool
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
	).WithRuntime(
		"custom-container",
	).WithCustomContainerConfig(
		fc.NewCustomContainerConfig().WithImage(
			functionConfig.ContainerImage,
		),
	)
	if functionConfig.UpdateEnvironmentVariables {
		request = request.WithEnvironmentVariables(
			functionConfig.EnvironmentVariables)
	}
	if !functionConfig.Yes {
		if !common.ComfirmDeploy() {
			return nil, common.ErrCanceled
		}
	}
	output, err := client.UpdateFunction(request)
	return output, err
}

func _getRegionFromRepository(repository string) (string, error) {
	// repository example: registry.cn-zhangjiakou.aliyuncs.com/space/name
	parts := strings.SplitN(repository, ".", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("no region in repository %s", repository)
	}
	return parts[1], nil
}

type DeployParams struct {
	ServiceName          string
	FunctionName         string
	Repository           string
	BuildId              string
	EnvironmentVariables *map[string]string
	Yes                  bool
}

func DoDeploy(params DeployParams) (*fc.UpdateFunctionOutput, error) {
	accessConfig, err := LoadAccessConfig()
	if err != nil {
		return nil, err
	}
	containerImage := fmt.Sprintf("%s:%s", params.Repository, params.BuildId)
	region, err := _getRegionFromRepository(params.Repository)
	if err != nil {
		return nil, err
	}
	functionConfig := _FunctionConfig{
		Region:                     region,
		FunctionName:               params.FunctionName,
		ServiceName:                params.ServiceName,
		ContainerImage:             containerImage,
		UpdateEnvironmentVariables: params.EnvironmentVariables != nil,
		EnvironmentVariables:       *params.EnvironmentVariables,
		Yes:                        params.Yes,
	}
	output, err := _updateFunction(accessConfig, &functionConfig)
	if err != nil {
		return nil, err
	}
	return output, nil
}
