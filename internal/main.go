package internal

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

func _MakeDeployCommand() *cobra.Command {
	var params DeployParams
	deploy := cobra.Command{
		Use:   "deploy",
		Short: "deploy function",
		Run: func(cmd *cobra.Command, args []string) {
			DoDeploy(params)
		},
	}
	deploy.Flags().StringVar(
		&params.ServiceName, "service", "", "service name [required]")
	deploy.MarkFlagRequired("service")
	deploy.Flags().StringVar(
		&params.FunctionName, "function", "", "function name [required]")
	deploy.MarkFlagRequired("function")
	deploy.Flags().StringVar(
		&params.ContainerImage, "image", "", "container image [required]")
	deploy.MarkFlagRequired("image")
	deploy.Flags().StringVar(
		&params.Envfile, "envfile", "", "envfile path")
	return &deploy
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
	err := cli.Execute()
	if err != nil {
		os.Exit(1)
	}
}
