package common

import (
	"fmt"
	"os"
	"os/exec"
	gofilepath "path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
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
	File         string
	Path         string
	ImageList    []string
	Platform     string
	Progress     string
	BuildArgList []string
}

func _getPlatform(p DockerBuildParams) string {
	platform := p.Platform
	if platform == "" {
		platform = "linux/amd64"
	}
	return platform
}

func _getProgress(p DockerBuildParams) string {
	progress := p.Progress
	if progress == "" {
		progress = "auto"
	}
	return progress
}

/* Call docker build command */
func DockerBuild(p DockerBuildParams) error {
	var commandArgs []string = []string{
		"build",
		"--platform", _getPlatform(p),
		"--progress", _getProgress(p),
		"-f", p.File,
	}
	for _, image := range p.ImageList {
		commandArgs = append(commandArgs, "-t", image)
	}
	for _, arg := range p.BuildArgList {
		commandArgs = append(commandArgs, "--build-arg")
		commandArgs = append(commandArgs, arg)
	}
	commandArgs = append(commandArgs, p.Path)
	return Shell("docker", commandArgs...)
}

func isFileExecAny(filepath string) bool {
	fileinfo, err := os.Stat(filepath)
	if err != nil {
		return false
	}
	mode := fileinfo.Mode()
	return mode&0111 != 0
}

func DockerScriptBuild(script string, p DockerBuildParams) error {
	script, err := homedir.Expand(script)
	if err != nil {
		return err
	}
	script = gofilepath.ToSlash(script)
	var envList []string = os.Environ()
	envList = append(envList, []string{
		fmt.Sprintf("EZFAAS_BUILD_PLATFORM=%s", _getPlatform(p)),
		fmt.Sprintf("EZFAAS_BUILD_PROGRESS=%s", _getProgress(p)),
		fmt.Sprintf("EZFAAS_BUILD_DOCKER_FILE=%s", p.File),
		fmt.Sprintf("EZFAAS_BUILD_DOCKER_IMAGE=%s", p.ImageList[0]),
	}...)
	// https://docs.docker.com/engine/reference/commandline/build/#set-build-time-variables---build-arg
	for _, item := range p.BuildArgList {
		parts := strings.SplitN(item, "=", 2)
		// ignore if not key=value format, not need to process
		if len(parts) == 2 {
			k, v := parts[0], parts[1]
			envList = append(envList, fmt.Sprintf("%s=%s", k, v))
		}
	}
	var cmd *exec.Cmd
	if isFileExecAny(script) {
		cmd = exec.Command(script)
	} else {
		cmd = exec.Command("bash", script)
	}
	cmd.Env = envList
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmdErr := cmd.Run()
	return cmdErr
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
