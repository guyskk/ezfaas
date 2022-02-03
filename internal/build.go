package internal

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/guyskk/ezfaas/internal/common"
)

func _randomHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func GetBuildId(suffix string) string {
	now := time.Now().UTC()
	return fmt.Sprintf(
		"%04d%02d%02d-%02d%02d%02d-%s",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		suffix,
	)
}

type BuildParams struct {
	Dockerfile string
	Path       string
	Repository string
}

type BuildResult struct {
	BuildId  string
	CommitId string
	Image    string
}

func Build(p BuildParams) (*BuildResult, error) {
	var suffix string
	commitId, err := common.GetCommitId()
	if err != nil {
		log.Printf("[WARN] %s", err)
		suffix = _randomHex(2)
	} else {
		suffix = commitId[:6]
	}
	buildId := GetBuildId(suffix)
	image := fmt.Sprintf("%s:%s", p.Repository, buildId)
	var buildArgs = map[string]string{
		"EZFAAS_COMMIT_ID": commitId,
		"EZFAAS_BUILD_ID":  buildId,
	}
	log.Printf("[INFO] COMMIT_ID=%s", commitId)
	log.Printf("[INFO] BUILD_ID=%s", buildId)
	log.Printf("[INFO] IMAGE=%s", image)
	buildParams := common.DockerBuildParams{
		File:      p.Dockerfile,
		Path:      p.Path,
		Image:     image,
		BuildArgs: buildArgs,
	}
	buildErr := common.DockerBuild(buildParams)
	if buildErr != nil {
		return nil, buildErr
	}
	result := BuildResult{
		BuildId:  buildId,
		CommitId: commitId,
		Image:    image,
	}
	return &result, nil
}

func DoBuild(p BuildParams) {
	_, err := Build(p)
	if err != nil {
		log.Fatal(err)
	}
}
