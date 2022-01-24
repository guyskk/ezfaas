package internal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func Shell(name string, arg ...string) error {
	// https://stackoverflow.com/questions/8875038/redirect-stdout-pipe-of-child-process-in-go
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

/* Get current git commit id, return "" if not available */
func GetCommitId() string {
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Get commit id failed: %s", err)
		return ""
	}
	return strings.TrimSpace(string(output))
}

type DockerBuildParams = struct {
	File      string
	Path      string
	Image     string
	BuildArgs map[string]string
}

/* Call docker build command */
func DockerBuild(p DockerBuildParams) error {
	var commandArgs []string = []string{
		"build",
		"--platform", "linux/amd64",
		"-f", p.File,
		"-t", p.Image,
	}
	for k, v := range p.BuildArgs {
		commandArgs = append(commandArgs, "--build-arg")
		commandArgs = append(commandArgs, fmt.Sprintf("%s=%s", k, v))
	}
	commandArgs = append(commandArgs, p.Path)
	return Shell("docker", commandArgs...)
}

type DockerPushParams struct {
	Image string
}

/* Call docker push command */
func DockerPush(p DockerPushParams) error {
	return Shell("docker", "push", p.Image)
}
