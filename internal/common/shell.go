package common

import (
	"fmt"
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

/* Get current git commit id */
func GetCommitId() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("get commit id failed: %s", err)
	}
	return strings.TrimSpace(string(output)), nil
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

/* Get docker image digest value */
func GetDockerImageDigest(image string) (string, error) {
	outFormat := "{{index .RepoDigests 0}}"
	cmd := exec.Command("docker", "image", "inspect", image, "--format", outFormat)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("get docker image digest failed: %s", err)
	}
	// example: ccr.ccs.tencentyun.com/ezfuns/shopant@sha256:1391376a56dexxx
	outputStr := strings.TrimSpace(string(output))
	parts := strings.SplitAfter(outputStr, "@")
	if len(parts) < 2 {
		return "", fmt.Errorf("not found docker image digest %s", outputStr)
	}
	return parts[1], nil
}

type DockerPushParams struct {
	Image string
}

/* Call docker push command */
func DockerPush(p DockerPushParams) error {
	return Shell("docker", "push", p.Image)
}
