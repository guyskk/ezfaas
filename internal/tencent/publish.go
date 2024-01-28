package tencent

import (
	"log"

	scf "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/scf/v20180416"
)

func _getFunctionVersionList(client *scf.Client, functionName string) ([]*scf.FunctionVersion, error) {
	request := scf.NewListVersionByFunctionRequest()
	request.FunctionName = &functionName
	response, err := client.ListVersionByFunction(request)
	if err != nil {
		return nil, err
	}
	return response.Response.Versions, nil
}

func _getFunctionAliasList(client *scf.Client, functionName string) ([]*scf.Alias, error) {
	request := scf.NewListAliasesRequest()
	request.FunctionName = &functionName
	response, err := client.ListAliases(request)
	if err != nil {
		return nil, err
	}
	return response.Response.Aliases, nil
}

func _isIntegerString(str string) bool {
	for _, c := range str {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func _getUsedAliasVersion(alias *scf.Alias) *string {
	if alias.FunctionVersion == nil {
		return nil
	}
	version := *alias.FunctionVersion
	if version == "$LATEST" {
		return &version
	}
	if alias.Name != nil && *alias.Name == "$DEFAULT" {
		return &version
	}
	if alias.RoutingConfig != nil {
		weights := alias.RoutingConfig.AdditionalVersionWeights
		if len(weights) > 0 {
			return &version
		}
		matchs := alias.RoutingConfig.AddtionVersionMatchs
		if len(matchs) > 0 {
			return &version
		}
	}
	return nil
}

func _deleteFunctionVersion(client *scf.Client, functionName string, version string) error {
	request := scf.NewDeleteFunctionRequest()
	request.FunctionName = &functionName
	request.Qualifier = &version
	_, err := client.DeleteFunction(request)
	return err
}

// 删除没有流量的数字版本
func _doDeleteOldVersion(client *scf.Client,
	params DeployParams) error {
	funcName := params.FunctionName
	versionList, versionErr := _getFunctionVersionList(client, funcName)
	if versionErr != nil {
		return versionErr
	}
	aliasList, aliasErr := _getFunctionAliasList(client, funcName)
	if aliasErr != nil {
		return aliasErr
	}
	// 将数字版本标记为删除
	deleteVersionMap := make(map[string]bool)
	if len(versionList) > 1 {
		for _, version := range versionList {
			if version.Version == nil {
				continue
			}
			if _isIntegerString(*version.Version) {
				deleteVersionMap[*version.Version] = true
			}
		}
	}
	// 将有流量的版本标记为不删除
	for _, alias := range aliasList {
		version := _getUsedAliasVersion(alias)
		if version != nil {
			deleteVersionMap[*version] = false
		}
	}
	// 删除标记为删除的版本
	for _, version := range versionList {
		if version.Version == nil {
			continue
		}
		if deleteVersionMap[*version.Version] {
			log.Printf("[INFO] Delete %s version %s", funcName, *version.Version)
			deleteErr := _deleteFunctionVersion(client, params.FunctionName, *version.Version)
			if deleteErr != nil {
				return deleteErr
			}
		}
	}
	return nil
}

func _doPublish(
	client *scf.Client,
	params DeployParams,
) (*scf.PublishVersionResponse, error) {
	request := scf.NewPublishVersionRequest()
	request.FunctionName = &params.FunctionName
	return client.PublishVersion(request)
}

// 未使用，代码保留备用
func DoPublish(client *scf.Client,
	params DeployParams) error {
	log.Println("[INFO] Publish function...")
	deleteErr := _doDeleteOldVersion(client, params)
	if deleteErr != nil {
		return deleteErr
	}
	_, publishErr := _doPublish(client, params)
	if publishErr != nil {
		return publishErr
	}
	return nil
}
