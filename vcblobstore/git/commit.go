package git

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	SimulateGitCommitFailureEnvvarName = "GIT_COMMIT_FAIL_INTRUSIVE_TEST"
	GitCommitFailureTestCommand        = "procyon lotor"
)

type CommitQueryResponseItem struct {
	Id             string `json:"id"`
	CommittedDate  string `json:"committed_date"`
	Message        string `json:"message"`
	AuthorName     string `json:"author_name"`
	AuthorEmail    string `json:"author_email"`
	AuthoredDate   string `json:"authored_date"`
	CommitterName  string `json:"committer_name"`
	CommitterEmail string `json:"committer_email"`
}

type CommitMetadata struct {
	Author     string
	AuthorDate time.Time
	Commit     string
	CommitDate time.Time
	Message    string
}

var (
	authorRegexp     = regexp.MustCompile(`^Author:[\s]+(.+)$`)
	authorDateRegexp = regexp.MustCompile(`^AuthorDate:[\s]+(.+)([0-9]{2})([0-9]{2})$`)
	commitRegexp     = regexp.MustCompile(`^Commit:[\s]+(.+)$`)
	commitDateRegexp = regexp.MustCompile(`^CommitDate:[\s]+(.+)([0-9]{2})([0-9]{2})$`)
)

func ParseLocalCommitMetadata(metadata string) (CommitMetadata, error) {
	commitMetadata := CommitMetadata{}
	commitMessageBuffer := []string{}

	lines := strings.Split(metadata, "\n")
	for _, line := range lines {
		submatch := authorRegexp.FindStringSubmatch(line)
		if submatch != nil {
			commitMetadata.Author = submatch[1]
			continue
		}

		submatch = commitRegexp.FindStringSubmatch(line)
		if submatch != nil {
			commitMetadata.Commit = submatch[1]
			continue
		}

		tim, found, err := parseTimeFromLocalCommitOutput(*authorDateRegexp, line)
		if err != nil {
			return commitMetadata, err
		}
		if found {
			commitMetadata.AuthorDate = tim
			continue
		}

		tim, found, err = parseTimeFromLocalCommitOutput(*commitDateRegexp, line)
		if err != nil {
			return commitMetadata, err
		}
		if found {
			commitMetadata.CommitDate = tim
			continue
		}

		if strings.Index(line, "    ") == 0 {
			commitMessageBuffer = append(commitMessageBuffer, strings.Trim(line, " \t"))
		}
	}

	commitMetadata.Message = strings.Join(commitMessageBuffer, "\n")

	return commitMetadata, nil
}

func parseTimeFromLocalCommitOutput(rexp regexp.Regexp, line string) (time.Time, bool, error) {
	if rexp.MatchString(line) {
		submatch := rexp.FindStringSubmatch(line)
		rfc3339 := fmt.Sprintf("%s%s:%s", submatch[1], submatch[2], submatch[3])
		t, err := time.Parse(time.RFC3339, rfc3339)
		if err != nil {
			return time.Time{}, false, fmt.Errorf("failed to parse time `%s` as RFC3339: %w", rfc3339, err)
		}
		return t, true, nil
	}
	return time.Time{}, false, nil
}

func GitlabCommitResponseToMetadata(response CommitQueryResponseItem) (CommitMetadata, error) {
	authorDate, err := time.Parse(time.RFC3339, response.AuthoredDate)
	if err != nil {
		return CommitMetadata{}, fmt.Errorf("failed to parse time `%s` as RFC3339: %w", response.AuthoredDate, err)
	}

	commitDate, err := time.Parse(time.RFC3339, response.CommittedDate)
	if err != nil {
		return CommitMetadata{}, fmt.Errorf("failed to parse time `%s` as RFC3339: %w", response.CommittedDate, err)
	}

	return CommitMetadata{
		Author:     fmt.Sprintf("%s <%s>", response.AuthorName, response.AuthorEmail),
		AuthorDate: authorDate,
		Commit:     fmt.Sprintf("%s <%s>", response.CommitterName, response.CommitterEmail),
		CommitDate: commitDate,
	}, nil
}
