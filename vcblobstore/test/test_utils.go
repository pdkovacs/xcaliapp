package test

import (
	"vcblobstore/git/gitlab"
	"context"
	"fmt"
	"os"
	"regexp"
)

const defaultGitlabProjectPath = "iconrepo-gitrepo-test"

const gitlabAPITokenLineRegexpString = "GITLAB_ACCESS_TOKEN=?(.+)"

var gitlabAPITokenLineRegexp = regexp.MustCompile(gitlabAPITokenLineRegexpString)

func GitTestGitlabAPIToken() (string, error) {
	homeDir, homedirErr := os.UserHomeDir()
	if homedirErr != nil {
		return "", fmt.Errorf("failed to get gitlab API token: %w", homedirErr)
	}
	content, readErr := os.ReadFile(fmt.Sprintf("%s/.iconrepo.secrets", homeDir))
	if readErr != nil {
		return "", fmt.Errorf("failed to get gitlab API token: %w", readErr)
	}

	hasMatch := gitlabAPITokenLineRegexp.Match(content)
	if !hasMatch {
		return "", fmt.Errorf("no match for gitlab api token pattern in content. I was looking for: %s", gitlabAPITokenLineRegexpString)
	}

	submatches := gitlabAPITokenLineRegexp.FindStringSubmatch(string(content))
	if len(submatches) < 2 {
		return "", fmt.Errorf("no match for gitlab api token pattern in content")
	}
	return submatches[1], nil
}

func SetupGitlabTestCaseConfig(conf *gitlab.Config, testSequenceId string, testCaseId string) {
	conf.GitlabProjectPath = fmt.Sprintf("%s_%s_%s", defaultGitlabProjectPath, testSequenceId, testCaseId)
}

func NewGitlabTestRepoClient(conf *gitlab.Config) (*gitlab.Gitlab, error) {
	conf.GitlabNamespacePath = "testing-with-repositories"

	var apiTokenErr error
	conf.GitlabAccessToken, apiTokenErr = GitTestGitlabAPIToken()
	if apiTokenErr != nil {
		return nil, apiTokenErr
	}

	gitlab, err := gitlab.NewGitlabRepositoryClient(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create gitlab repo client %w", err)
	}

	return gitlab, nil
}

type RepositoryResetter interface {
	ResetRepository() error
}
