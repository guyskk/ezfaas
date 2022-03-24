package tencent

import (
	"fmt"
	"log"
	"strings"
	"time"

	ezcommon "github.com/guyskk/ezfaas/internal/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	scf "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/scf/v20180416"
)

type DeployParams struct {
	Region               string
	FunctionName         string
	Repository           string
	BuildId              string
	EnvironmentVariables *map[string]string
	Yes                  bool
}

const (
	//	函数创建中
	FUNCTION_STATUS_CREATING string = "Creating"
	//	函数创建失败（如果已生成函数信息，则可以删除后再创建）
	FUNCTION_STATUS_CREATEFAILED string = "CreateFailed"
	//	函数可用
	FUNCTION_STATUS_ACTIVE string = "Active"
	//	函数更新中
	FUNCTION_STATUS_UPDATING string = "Updating"
	//	函数更新失败
	FUNCTION_STATUS_UPDATEFAILED string = "UpdateFailed"
	//	函数版本发布中
	FUNCTION_STATUS_PUBLISHING string = "Publishing"
	//	函数版本发布失败
	FUNCTION_STATUS_PUBLISHFAILED string = "PublishFailed"
	//	函数删除中
	FUNCTION_STATUS_DELETING string = "Deleting"
	//	函数删除失败
	FUNCTION_STATUS_DELETEFAILED string = "DeleteFailed"
)

func _isFailedStatus(status string) bool {
	return strings.Contains(status, "Failed")
}

func _getFunctionInfo(
	client *scf.Client,
	params DeployParams,
) (*scf.GetFunctionResponse, error) {
	request := scf.NewGetFunctionRequest()
	request.FunctionName = &params.FunctionName
	return client.GetFunction(request)
}

/*
   当使用 云 API 管理函数（增、删、改、查）时，由于接口异步，
   首先需要查询函数当前状态，然后进行下一步操作。
   https://cloud.tencent.com/document/product/583/47175
*/
func _getFunctionStatus(
	client *scf.Client,
	params DeployParams,
) (string, error) {
	response, err := _getFunctionInfo(client, params)
	if err != nil {
		return "", err
	}
	return *response.Response.Status, nil
}

func _waitFunctionActive(
	client *scf.Client,
	params DeployParams,
	timeout time.Duration,
) error {
	deadline := time.Now().Add(timeout)
	i := 1
	for {
		status, err := _getFunctionStatus(client, params)
		if err != nil {
			return err
		}
		if status == FUNCTION_STATUS_ACTIVE {
			return nil
		}
		if _isFailedStatus(status) {
			return fmt.Errorf("function failed, status=%s", status)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("function not active, status=%s", status)
		}
		if i%3 == 0 {
			log.Printf("[INFO] Wait function active, status=%s", status)
		}
		time.Sleep(time.Duration(1 * time.Second))
		i += 1
	}
}

func _updateCode(
	client *scf.Client,
	params DeployParams,
	imageUri string,
) (*scf.UpdateFunctionCodeResponse, error) {
	request := scf.NewUpdateFunctionCodeRequest()
	request.FunctionName = &params.FunctionName
	imageType := "personal"
	request.Code = &scf.Code{
		ImageConfig: &scf.ImageConfig{
			ImageType: &imageType,
			ImageUri:  &imageUri,
		},
	}
	return client.UpdateFunctionCode(request)
}

func _updateConfig(
	client *scf.Client,
	params DeployParams,
) (*scf.UpdateFunctionConfigurationResponse, error) {
	var Variables []*scf.Variable
	for k, v := range *params.EnvironmentVariables {
		key, value := k, v // 更新变量地址
		Variables = append(Variables, &scf.Variable{Key: &key, Value: &value})
	}
	request := scf.NewUpdateFunctionConfigurationRequest()
	request.FunctionName = &params.FunctionName
	request.Environment = &scf.Environment{Variables: Variables}
	return client.UpdateFunctionConfiguration(request)
}

func _doPublish(
	client *scf.Client,
	params DeployParams,
) (*scf.PublishVersionResponse, error) {
	request := scf.NewPublishVersionRequest()
	request.FunctionName = &params.FunctionName
	return client.PublishVersion(request)
}

func DoDeploy(params DeployParams) (*scf.GetFunctionResponse, error) {
	dockerImage := fmt.Sprintf("%s:%s", params.Repository, params.BuildId)
	imageDigest, digestErr := ezcommon.GetDockerImageDigest(dockerImage)
	if digestErr != nil {
		return nil, digestErr
	}
	imageUri := fmt.Sprintf("%s@%s", dockerImage, imageDigest)
	hasEnvironmentVariables := params.EnvironmentVariables != nil
	log.Printf("[INFO] Region=%s Function=%s", params.Region, params.FunctionName)
	log.Printf("[INFO] ContainerImage=%s", dockerImage)
	log.Printf("[INFO] ContainerImageDigest=%s", imageDigest)
	log.Printf("[INFO] UpdateEnvironmentVariables=%t", hasEnvironmentVariables)
	provider := common.DefaultProfileProvider()
	credentail, err := provider.GetCredential()
	if err != nil {
		return nil, err
	}
	clientProfile := profile.NewClientProfile()
	client, err := scf.NewClient(credentail, params.Region, clientProfile)
	if err != nil {
		return nil, err
	}
	if !params.Yes {
		if !ezcommon.ComfirmDeploy() {
			return nil, ezcommon.ErrCanceled
		}
	}
	status, statusErr := _getFunctionStatus(client, params)
	if statusErr != nil {
		return nil, statusErr
	}
	if status != FUNCTION_STATUS_ACTIVE {
		return nil, fmt.Errorf("function not active, status=%s", status)
	}
	imageErr := WaitDockerImageReady(WaitDockerImageParams{
		Region:     params.Region,
		Repository: params.Repository,
		BuildId:    params.BuildId,
		Timeout:    time.Duration(30 * time.Second),
	})
	if imageErr != nil {
		return nil, imageErr
	}
	log.Println("[INFO] Update function code...")
	_, codeErr := _updateCode(client, params, imageUri)
	if codeErr != nil {
		return nil, codeErr
	}
	waitFunctionTimeout := time.Duration(30 * time.Second)
	err = _waitFunctionActive(client, params, waitFunctionTimeout)
	if err != nil {
		return nil, err
	}
	if hasEnvironmentVariables {
		log.Println("[INFO] Update function config...")
		_, configErr := _updateConfig(client, params)
		if configErr != nil {
			return nil, configErr
		}
		err = _waitFunctionActive(client, params, waitFunctionTimeout)
		if err != nil {
			return nil, err
		}
	}
	log.Println("[INFO] Publish function...")
	_, publishErr := _doPublish(client, params)
	if publishErr != nil {
		return nil, publishErr
	}
	response, err := _getFunctionInfo(client, params)
	if err != nil {
		return nil, err
	}
	return response, nil
}
