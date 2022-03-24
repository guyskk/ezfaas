package tencent

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tcr/v20190924"
)

type WaitDockerImageParams struct {
	Region     string
	Repository string
	BuildId    string
	Timeout    time.Duration
}

func extractRepoName(repository string) (string, error) {
	parts := strings.Split(repository, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid repository name")
	}
	parts = parts[len(parts)-2:]
	return parts[0] + "/" + parts[1], nil
}

func queryImageTagReady(client *tcr.Client, repoName string, tag string) (bool, error) {
	request := tcr.NewDescribeImagePersonalRequest()
	request.RepoName = strRef(repoName)
	request.Limit = int64Ref(1)
	request.Offset = int64Ref(0)
	request.Tag = strRef(tag)
	response, err := client.DescribeImagePersonal(request)
	if err != nil {
		return false, err
	}
	isReady := len(response.Response.Data.TagInfo) >= 1
	return isReady, nil
}

func WaitDockerImageReady(params WaitDockerImageParams) error {
	repoName, err := extractRepoName(params.Repository)
	if err != nil {
		return err
	}
	provider := common.DefaultProfileProvider()
	credentail, err := provider.GetCredential()
	if err != nil {
		return err
	}
	clientProfile := profile.NewClientProfile()
	client, err := tcr.NewClient(credentail, params.Region, clientProfile)
	if err != nil {
		return err
	}
	deadline := time.Now().Add(params.Timeout)
	i := 1
	for {
		isReady, err := queryImageTagReady(client, repoName, params.BuildId)
		if err != nil {
			return err
		}
		if isReady {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("docker image not ready")
		}
		if i%3 == 0 {
			log.Printf("[INFO] Wait docker image ready...")
		}
		time.Sleep(time.Duration(1 * time.Second))
		i += 1
	}
}
