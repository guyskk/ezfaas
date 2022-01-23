package internal

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

func _MakeDeployCommand() *cobra.Command {
	var params DeployParams
	cmd := cobra.Command{
		Use:   "deploy",
		Short: "deploy function",
		Run: func(cmd *cobra.Command, args []string) {
			DoDeploy(params)
		},
	}
	cmd.Flags().StringVar(
		&params.ServiceName, "service", "", "service name [required]")
	cmd.MarkFlagRequired("service")
	cmd.Flags().StringVar(
		&params.FunctionName, "function", "", "function name [required]")
	cmd.MarkFlagRequired("function")
	cmd.Flags().StringVar(
		&params.ContainerImage, "image", "", "container image [required]")
	cmd.MarkFlagRequired("image")
	cmd.Flags().StringVar(
		&params.Envfile, "envfile", "", "envfile path")
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
	cmd.Flags().StringVar(
		&params.Dockerfile, "dockerfile", "Dockerfile", "Dockerfile path")
	cmd.Flags().StringVar(
		&params.Path, "path", ".", "Docker build path")
	cmd.Flags().StringVar(
		&params.Repository, "repository", "", "docker image repository [required]")
	cmd.MarkFlagRequired("repository")
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
	cli.AddCommand(_MakeDeployCommand())
	cli.AddCommand(_MakeBuildCommand())
	err := cli.Execute()
	if err != nil {
		os.Exit(1)
	}
}
