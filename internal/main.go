package internal

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

func _AddBaseDeployFlags(cmd *cobra.Command, params *BaseDeployParams) {
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(
		&params.FunctionName, "function", "", "Function name [required]")
	cmd.MarkFlagRequired("function")
	cmd.Flags().StringVar(
		&params.Repository, "repository", "", "Docker image repository [required]")
	cmd.MarkFlagRequired("repository")
	cmd.Flags().StringVar(
		&params.BuildId, "build-id", "", "Existed build id (image version)")
	cmd.Flags().StringVar(
		&params.Dockerfile, "dockerfile", "Dockerfile", "Dockerfile path")
	cmd.Flags().StringVar(
		&params.BuildPath, "path", ".", "Docker build path")
	cmd.Flags().StringVar(
		&params.Envfile, "envfile", "", "Envfile path")
	cmd.Flags().BoolVar(
		&params.Yes, "yes", false, "Confirm deploy")
}

func _MakeDeployAliyunCommand() *cobra.Command {
	var params AliyunDeployParams
	cmd := cobra.Command{
		Use:   "deploy-aliyun",
		Short: "deploy function to aliyun",
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
		Short: "deploy function to tencent",
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
		Short: "build container image",
		Run: func(cmd *cobra.Command, args []string) {
			DoBuild(params)
		},
	}
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(
		&params.Dockerfile, "dockerfile", "Dockerfile", "Dockerfile path")
	cmd.Flags().StringVar(
		&params.Path, "path", ".", "Docker build path")
	cmd.Flags().StringVar(
		&params.Repository, "repository", "", "Docker image repository [required]")
	cmd.MarkFlagRequired("repository")
	return &cmd
}

func _MakeConfigCdnCacheTencentCommand() *cobra.Command {
	var params TencentCDNCacheConfigParams
	cmd := cobra.Command{
		Use:   "config-cdn-cache-tencent",
		Short: "config CDN cache rules tencent",
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
