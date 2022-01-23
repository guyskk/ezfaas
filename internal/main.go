package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	gofilepath "path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/aliyun/fc-go-sdk"
	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

type AccessConfig struct {
	ALIBABA_CLOUD_ACCOUNT_ID        string
	ALIBABA_CLOUD_ACCESS_KEY_ID     string
	ALIBABA_CLOUD_ACCESS_KEY_SECRET string
}

func readUserFile(filepath string) ([]byte, error) {
	filepath, err := homedir.Expand(filepath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(gofilepath.ToSlash(filepath))
}

func loadAccessConfig() (*AccessConfig, error) {
	data, err := readUserFile("~/.config/aliyun_fc_deploy.json")
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

type FunctionConfig struct {
	ServiceName                string
	FunctionName               string
	ContainerImage             string
	UpdateEnvironmentVariables bool
	EnvironmentVariables       map[string]string
}

func updateFunction(
	accessConfig *AccessConfig,
	functionConfig *FunctionConfig,
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

func onDeploy(
	serviceName string,
	functionName string,
	containerImage string,
	envfile string,
) {
	accessConfig, err := loadAccessConfig()
	if err != nil {
		log.Fatal(err)
	}
	tomlLoads := toml.Unmarshal
	if tomlLoads == nil {
		fmt.Println("")
	}
	var env map[string]string
	if envfile != "" {
		envdata, err := readUserFile(envfile)
		if err != nil {
			log.Fatal(err)
		}
		env, err = godotenv.Unmarshal(string(envdata))
		if err != nil {
			log.Fatal(err)
		}
	}
	functionConfig := FunctionConfig{
		FunctionName:               functionName,
		ServiceName:                serviceName,
		ContainerImage:             containerImage,
		UpdateEnvironmentVariables: envfile != "",
		EnvironmentVariables:       env,
	}
	output, err := updateFunction(accessConfig, &functionConfig)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s \n", output)
}

func main() {
	cli := cobra.Command{
		Use:   "ezfaas",
		Short: "EZ FaaS Toolkit",
		Run: func(cmd *cobra.Command, args []string) {
			err := cmd.Help()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	var serviceName string
	var functionName string
	var containerImage string
	var envfile string
	deploy := cobra.Command{
		Use:   "deploy",
		Short: "deploy function",
		Run: func(cmd *cobra.Command, args []string) {
			onDeploy(
				serviceName,
				functionName,
				containerImage,
				envfile,
			)
		},
	}
	deploy.Flags().StringVar(
		&serviceName, "service", "", "service name [required]")
	deploy.MarkFlagRequired("service")
	deploy.Flags().StringVar(
		&functionName, "function", "", "function name [required]")
	deploy.MarkFlagRequired("function")
	deploy.Flags().StringVar(
		&containerImage, "image", "", "container image [required]")
	deploy.MarkFlagRequired("image")
	deploy.Flags().StringVar(
		&envfile, "envfile", "", "envfile path")
	cli.AddCommand(&deploy)
	err := cli.Execute()
	if err != nil {
		os.Exit(1)
	}
}
