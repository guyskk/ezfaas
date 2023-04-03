package internal

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

func _AddBaseBuildFlags(cmd *cobra.Command, params *BaseBuildParams) {
	cmd.Flags().StringVar(
		&params.Dockerfile, "dockerfile", "Dockerfile", "Dockerfile path")
	cmd.Flags().StringVar(
		&params.DockerConfig, "docker-config", "", "Docker config path")
	cmd.Flags().StringVar(
		&params.BuildPath, "build-path", ".", "Docker build path")
	cmd.Flags().StringVar(
		&params.BuildPlatform, "build-platform", "", "Docker build --platform")
	cmd.Flags().StringVar(
		&params.BuildProgress, "build-progress", "", "Docker build --progress")
	cmd.Flags().StringArrayVar(
		&params.ImageTagList, "image-tag", []string{}, "Docker build --tag name:version")
	cmd.Flags().StringArrayVar(
		&params.BuildArgList, "build-arg", []string{}, "Docker build --build-arg")
	cmd.Flags().StringVar(
		&params.BuildScript, "build-script", "", "Bash or executable script to build docker image")
}

func _AddBaseDeployFlags(cmd *cobra.Command, params *BaseDeployParams) {
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(
		&params.FunctionName, "function", "", "Function name [required]")
	cmd.MarkFlagRequired("function")
	cmd.Flags().StringVar(
		&params.Repository, "repository", "", "Docker image repository [required]")
	cmd.MarkFlagRequired("repository")
	_AddBaseBuildFlags(cmd, &params.BaseBuildParams)
	cmd.Flags().StringVar(
		&params.BuildId, "build-id", "", "Existed build id (image version)")
	cmd.Flags().StringVar(
		&params.Envfile, "envfile", "", "Envfile path")
	cmd.Flags().BoolVar(
		&params.Yes, "yes", false, "Confirm deploy")
}

func _MakeDeployAliyunCommand() *cobra.Command {
	var params AliyunDeployParams
	cmd := cobra.Command{
		Use:   "deploy-aliyun",
		Short: "Deploy function to aliyun",
		Run: func(cmd *cobra.Command, args []string) {
			DoDeployAliyun(params)
		},
	}
	cmd.Flags().StringVar(
		&params.ServiceName, "service", "", "Service name [required]")
	cmd.MarkFlagRequired("service")
	_AddBaseDeployFlags(&cmd, &params.BaseDeployParams)
	return &cmd
}

func _MakeDeployTencentCommand() *cobra.Command {
	var params TencentDeployParams
	cmd := cobra.Command{
		Use:   "deploy-tencent",
		Short: "Deploy function to tencent",
		Run: func(cmd *cobra.Command, args []string) {
			DoDeployTencent(params)
		},
	}
	cmd.Flags().StringVar(
		&params.Region, "region", "", "Region name [required]")
	cmd.MarkFlagRequired("region")
	_AddBaseDeployFlags(&cmd, &params.BaseDeployParams)
	return &cmd
}

func _MakeBuildCommand() *cobra.Command {
	var params BuildParams
	cmd := cobra.Command{
		Use:   "build",
		Short: "Build docker image",
		Run: func(cmd *cobra.Command, args []string) {
			DoBuild(params)
		},
	}
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(
		&params.Repository, "repository", "", "Docker image repository [required]")
	cmd.MarkFlagRequired("repository")
	_AddBaseBuildFlags(&cmd, &params.BaseBuildParams)
	return &cmd
}

func _MakeConfigCdnCacheTencentCommand() *cobra.Command {
	var params TencentCDNCacheConfigParams
	cmd := cobra.Command{
		Use:   "config-cdn-cache-tencent",
		Short: "Config CDN cache rules of tencent",
		Run: func(cmd *cobra.Command, args []string) {
			DoConfigCdnCacheTencent(params)
		},
	}
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(
		&params.Region, "region", "", "Region name [required]")
	cmd.MarkFlagRequired("region")
	cmd.Flags().StringVar(
		&params.Domain, "domain", "", "Domain name [required]")
	cmd.MarkFlagRequired("domain")
	return &cmd
}

func Main() {
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
	cli.AddCommand(_MakeDeployAliyunCommand())
	cli.AddCommand(_MakeDeployTencentCommand())
	cli.AddCommand(_MakeBuildCommand())
	cli.AddCommand(_MakeConfigCdnCacheTencentCommand())
	err := cli.Execute()
	if err != nil {
		os.Exit(1)
	}
}
