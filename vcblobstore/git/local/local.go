package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"vcblobstore"
	"vcblobstore/git"
	"vcblobstore/git/local/config"

	"github.com/rs/zerolog"
)

const cleanStatusMessageTail = "nothing to commit, working tree clean"

type Git struct {
	location string
	logger   *zerolog.Logger
}

func (repo Git) String() string {
	return fmt.Sprintf("Local git repository at %s", repo.location)
}

func (repo *Git) CreateRepository(ctx context.Context) error {
	return repo.initMaybe()
}

func (repo *Git) ResetRepository(ctx context.Context) error {
	deleteRepoErr := repo.DeleteRepository(ctx)
	if deleteRepoErr != nil {
		panic(deleteRepoErr)
	}
	return repo.CreateRepository(ctx)
}

func (repo *Git) DeleteRepository(ctx context.Context) error {
	return os.RemoveAll(repo.location)
}

func (repo *Git) ExecuteGitCommand(args []string) (string, error) {
	return ExecuteCommand(ExecCmdParams{
		Name: "git",
		Args: args,
		Opts: &CmdOpts{Cwd: repo.location},
	}, repo.logger)
}

type gitJobMessages struct {
	logContext    string
	commitMessage string
}

func getCommitCommand() string {
	if os.Getenv(git.SimulateGitCommitFailureEnvvarName) == "true" {
		return git.GitCommitFailureTestCommand
	} else {
		return "commit"
	}
}

func commit(messageBase string, userName string) []string {
	return []string{
		getCommitCommand(),
		"-m", messageBase + " by " + userName,
		fmt.Sprintf("--author=%s <%s>", userName, userName),
	}
}

var rollbackCommands = [][]string{
	{"reset", "--hard", "HEAD"},
	{"clean", "-qfdx"},
}

func (repo *Git) rollback() {
	for _, rollbackCmd := range rollbackCommands {
		_, _ = repo.ExecuteGitCommand(rollbackCmd)
	}
}

func (repo *Git) executeBlobManipulationJob(blobOperation func() error, messages gitJobMessages, userName string) error {
	logger := repo.logger.With().Str("method", fmt.Sprintf("git: %s", messages.logContext)).Logger()

	if len(userName) == 0 {
		logger.Warn().Msg("Modifying user is not specified")
	}

	var out string
	var err error

	defer func() {
		if err != nil {
			logger.Debug().Err(err).Str("out", out).Msg("failed GIT operation")
			repo.rollback()
		} else {
			logger.Debug().Msg("Success")
		}
	}()

	err = blobOperation()
	if err != err {
		return fmt.Errorf("failed blob operation: %w", err)
	}
	out, err = repo.ExecuteGitCommand([]string{"add", "-A"})
	if err != nil {
		return fmt.Errorf("failed to add files to index: %w -> %s", err, out)
	}

	commitMessage := messages.commitMessage
	out, err = repo.ExecuteGitCommand(commit(commitMessage, userName))
	if err != nil {
		return fmt.Errorf("failed to commit: %w -> %s", err, out)
	}

	return err
}

func (repo *Git) createBlob(key string, content []byte) error {
	var err error

	logOp := func(opmsg string) string {
		repo.logger.Debug().Str("operation", opmsg).Msg("operation starting")
		return opmsg
	}

	path, pathErr := repo.pathToFile(key)
	if pathErr != nil {
		return pathErr
	}
	directory := filepath.Dir(path)

	operationMsg := logOp(fmt.Sprintf("create directory %s", directory))
	err = os.MkdirAll(directory, 0700)
	if err == nil {
		operationMsg = logOp(fmt.Sprintf("create directory %s", directory))
		err = os.MkdirAll(directory, 0700)
		if err == nil {
			operationMsg = logOp(fmt.Sprintf("write file %s", directory))
			err = os.WriteFile(path, content, 0700)
		}
	}
	if err != nil {
		err = fmt.Errorf("%s: %w", operationMsg, err)
	}
	return err
}

func (repo *Git) AddBlob(ctx context.Context, blob vcblobstore.BlobInfo) error {
	key := blob.Key
	content := blob.Content

	path, pathErr := repo.pathToFile(key)
	if pathErr != nil {
		return pathErr
	}

	blobOperation := func() error {
		err := repo.createBlob(key, content)
		if err != nil {
			return fmt.Errorf("failed to create blobfile %s as %s: %w", key, path, err)
		}
		return nil
	}

	jobTextProvider := gitJobMessages{
		"add blob file",
		"blob file version added",
	}

	var err error
	Enqueue(func() {
		err = repo.executeBlobManipulationJob(blobOperation, jobTextProvider, blob.ModifiedBy)
	})

	if err != nil {
		return fmt.Errorf("failed to add blobfile %v to git repository at %s: %w", path, repo.location, err)
	}
	return nil
}

func copyBlobContents(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	err = out.Sync()
	return err
}

func (repo *Git) CopyBlob(_ context.Context, sourceKey string, destinationKey string, modifiedBy string) error {
	jobTextProvider := gitJobMessages{
		"copy blob file",
		"blob file version added",
	}

	sourcePath, srcPathErr := repo.pathToFile(sourceKey)
	if srcPathErr != nil {
		return srcPathErr
	}
	destinationPath, destPathErr := repo.pathToFile(destinationKey)
	if destPathErr != nil {
		return destPathErr
	}

	blobOperation := func() error {
		err := copyBlobContents(sourcePath, destinationPath)
		if err != nil {
			return fmt.Errorf("failed to copy file contents from %s to %s: %w", sourceKey, destinationKey, err)
		}
		return nil
	}

	var err error
	Enqueue(func() {
		err = repo.executeBlobManipulationJob(blobOperation, jobTextProvider, modifiedBy)
	})

	if err != nil {
		repo.logger.Debug().Err(err).Msg("executeBlobManipulationJob failed while copying blob")
		return fmt.Errorf("failed to copy blobfile from %s to %s to git repository at %s: %w", sourceKey, destinationKey, repo.location, err)
	}
	return nil
}

func (repo *Git) GetBlob(ctx context.Context, key string) ([]byte, error) {
	path, pathErr := repo.pathToFile(key)
	if pathErr != nil {
		return nil, pathErr
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s from local git repo: %w", path, err)
	}
	return bytes, nil
}

func (repo *Git) deleteBlob(key string) error {
	path, pathErr := repo.pathToFile(key)
	if pathErr != nil {
		return pathErr
	}

	removeFileErr := os.Remove(path)
	if removeFileErr != nil {
		if removeFileErr == os.ErrNotExist {
			return fmt.Errorf("failed to remove blob %s: %w", key, vcblobstore.ErrBlobNotFound)
		}
		return fmt.Errorf("failed to remove blob %s: %w", key, removeFileErr)
	}
	return nil
}

func (repo *Git) DeleteBlob(ctx context.Context, key string, modifiedBy string) error {
	blobOperation := func() error {
		deletionError := repo.deleteBlob(key)
		return deletionError
	}

	jobTextProvider := gitJobMessages{
		fmt.Sprintf("delete blob %s", key),
		"blob deleted",
	}

	var err error
	Enqueue(func() {
		err = repo.executeBlobManipulationJob(blobOperation, jobTextProvider, modifiedBy)
	})

	if err != nil {
		return fmt.Errorf("failed to remove blob %s from git repository: %w", key, err)
	}
	return nil
}

func (repo Git) CheckStatus() (bool, error) {
	out, err := repo.ExecuteGitCommand([]string{"status"})
	if err != nil {
		return false, fmt.Errorf("failed to get current git commit: %w", err)
	}
	status := strings.TrimSpace(out)
	return strings.Contains(status, cleanStatusMessageTail), nil
}

func (repo Git) GetStateID(ctx context.Context) (string, error) {
	out, err := repo.ExecuteGitCommand([]string{"rev-parse", "HEAD"})
	if err != nil {
		return "", fmt.Errorf("failed to get current git commit: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func (repo Git) ListBlobKeys(ctx context.Context) ([]string, error) {
	output, err := repo.ExecuteGitCommand([]string{"ls-tree", "-r", "HEAD", "--name-only"})
	if err != nil {
		return nil, err
	}

	fileList := []string{}
	outputLines := strings.Split(output, config.LineBreak)
	for _, line := range outputLines {
		trimmedLine := strings.TrimSpace(line)
		if len(trimmedLine) > 0 {
			fileList = append(fileList, trimmedLine)
		}
	}
	return fileList, nil
}

// GetVersionFor returns the commit ID of the blob specified by the method paramters.
// Return empty string in case the file doesn't exist in the repository
func (repo Git) GetVersionFor(ctx context.Context, key string) (string, error) {
	path, pathErr := repo.pathToFile(key)
	if pathErr != nil {
		return "", pathErr
	}

	printCommitIDArgs := []string{"log", "-n", "1", "--pretty=format:%H", "--", path}
	output, execErr := repo.ExecuteGitCommand(printCommitIDArgs)
	if execErr != nil {
		return "", fmt.Errorf("failed to execute command to get last commit modifying %s: %w", key, execErr)
	}
	return output, nil
}

func (repo Git) GetVersionMetadata(ctx context.Context, commitId string) (git.CommitMetadata, error) {
	logger := repo.logger.With().Str("method", fmt.Sprintf("git: GetVersionMetadata: %s", commitId)).Logger()

	printCommitMetadataArgs := []string{"show", "--quiet", "--format=fuller", "--date=format:%Y-%m-%dT%H:%M:%S%z"}
	output, execErr := repo.ExecuteGitCommand(printCommitMetadataArgs)
	if execErr != nil {
		return git.CommitMetadata{}, fmt.Errorf("failed to get metadata from repo for commit %s: %w", commitId, execErr)
	}
	logger.Debug().Str("meta-data", output).Msg("raw metadata extracted")
	commitMetadata, parseErr := git.ParseLocalCommitMetadata(output)
	if parseErr != nil {
		return commitMetadata, fmt.Errorf("failed to parse metadata from commit %s: %w", commitId, parseErr)
	}
	return commitMetadata, nil
}

func (repo Git) createInitializeGitRepo() error {
	var err error
	var out string

	cmds := []ExecCmdParams{
		{Name: "rm", Args: []string{"-rf", repo.location}, Opts: nil},
		{Name: "mkdir", Args: []string{"-p", repo.location}, Opts: nil},
		{Name: "git", Args: []string{"init"}, Opts: &CmdOpts{Cwd: repo.location}},
	}

	for _, cmd := range cmds {
		out, err = ExecuteCommand(cmd, repo.logger)
		println(out)
		if err != nil {
			return fmt.Errorf("failed to create git repo at %s: %w", repo.location, err)
		}
	}

	return nil
}

func (repo Git) LocationHasRepo() bool {
	if GitRepoLocationExists(repo.location) {
		testCommand := ExecCmdParams{Name: "git", Args: []string{"init"}, Opts: &CmdOpts{Cwd: repo.location}}
		outOrErr, err := ExecuteCommand(testCommand, repo.logger)
		if err != nil {
			if strings.Contains(outOrErr, "not a git repository") { // TODO: Is it really possible to get this error message here?
				return false
			}
			panic(err)
		}
		return true
	}
	return false
}

// Init initializes the Git repository if it already doesn't exist
func (repo Git) initMaybe() error {
	if !repo.LocationHasRepo() {
		return repo.createInitializeGitRepo()
	}
	return nil
}

func (repo *Git) pathToFile(key string) (string, error) {
	if strings.Contains(key, "/") {
		return "", fmt.Errorf("invalid character in key: forward slash ('/')")
	}
	return filepath.Join(repo.location, key), nil
}

func GitRepoLocationExists(location string) bool {
	var err error
	var fi os.FileInfo

	fi, err = os.Stat(location)
	if err != nil {
		if os.IsNotExist(err) {
			// nothing to do here
			return false
		}
		panic(err)
	}

	if fi.Mode().IsRegular() {
		panic(fmt.Errorf("file exists, but it is not a directory: %s", location))
	}

	return true
}

type Config struct {
	Location string
}

func NewLocalGitRepository(localConfig *Config, logger *zerolog.Logger) *Git {
	git := Git{
		location: localConfig.Location,
		logger:   logger,
	}
	return &git
}
