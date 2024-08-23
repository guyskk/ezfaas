package aliyun

import (
	"fmt"
	"log"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	fc "github.com/alibabacloud-go/fc-20230330/v4/client"
	tea "github.com/alibabacloud-go/tea/tea"

	"github.com/guyskk/ezfaas/internal/common"
)

type _FunctionConfig struct {
	Region                     string
	FunctionName               string
	ContainerImage             string
	UpdateEnvironmentVariables bool
	EnvironmentVariables       map[string]string
	Yes                        bool
}

func _getEndpoint(accountId string, region string) string {
	endpoint := fmt.Sprintf(
		"%s.%s.fc.aliyuncs.com",
		accountId,
		region,
	)
	return endpoint
}

func _getClientConfig(accessConfig *AccessConfig, endpoint string) *openapi.Config {
	config := &openapi.Config{
		AccessKeyId:     tea.String(accessConfig.ALIBABA_CLOUD_ACCESS_KEY_ID),
		AccessKeySecret: tea.String(accessConfig.ALIBABA_CLOUD_ACCESS_KEY_SECRET),
		Endpoint:        tea.String(endpoint),
	}
	return config
}

func _updateFunction(
	accessConfig *AccessConfig,
	functionConfig *_FunctionConfig,
) (*fc.UpdateFunctionResponse, error) {
	log.Printf(
		"[INFO] Function=%s",
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
	endpoint := _getEndpoint(
		accessConfig.ALIBABA_CLOUD_ACCOUNT_ID,
		functionConfig.Region,
	)
	log.Printf("[INFO] Deploy Endpoint=%s", endpoint)
	clientConfig := _getClientConfig(accessConfig, endpoint)
	client, err := fc.NewClient(clientConfig)
	if err != nil {
		return nil, err
	}
	updateFunctionInput := fc.UpdateFunctionInput{
		CustomContainerConfig: &fc.CustomContainerConfig{
			Image: tea.String(functionConfig.ContainerImage),
		},
	}
	if functionConfig.UpdateEnvironmentVariables {
		fcEnvVars := map[string]*string{}
		for k, v := range functionConfig.EnvironmentVariables {
			fcEnvVars[k] = tea.String(v)
		}
		updateFunctionInput.EnvironmentVariables = fcEnvVars
	}
	request := fc.UpdateFunctionRequest{
		Body: &updateFunctionInput,
	}
	if !functionConfig.Yes {
		if !common.ComfirmDeploy() {
			return nil, common.ErrCanceled
		}
	}
	output, err := client.UpdateFunction(&functionConfig.FunctionName, &request)
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
	FunctionName         string
	Repository           string
	BuildId              string
	EnvironmentVariables *map[string]string
	Yes                  bool
}

func DoDeploy(params DeployParams) (*fc.UpdateFunctionResponse, error) {
	accessConfig, err := LoadAccessConfig()
	if err != nil {
		return nil, err
	}
	containerImage := fmt.Sprintf("%s:%s", params.Repository, params.BuildId)
	region, err := _getRegionFromRepository(params.Repository)
	if err != nil {
		return nil, err
	}
	hasEnv := params.EnvironmentVariables != nil
	env := map[string]string{}
	if hasEnv {
		env = *params.EnvironmentVariables
	}
	functionConfig := _FunctionConfig{
		Region:                     region,
		FunctionName:               params.FunctionName,
		ContainerImage:             containerImage,
		UpdateEnvironmentVariables: hasEnv,
		EnvironmentVariables:       env,
		Yes:                        params.Yes,
	}
	output, err := _updateFunction(accessConfig, &functionConfig)
	if err != nil {
		return nil, err
	}
	return output, nil
}
