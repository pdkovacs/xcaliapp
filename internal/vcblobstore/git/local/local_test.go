package local

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
)

var localTestConfig = &Config{
	Location: filepath.Join(os.Getenv("HOME"), "tmp", "versioned-vcblobstore", "gitrepo"),
}

func removeRepoDir() {
	rmdirErr := os.RemoveAll(localTestConfig.Location)
	if rmdirErr != nil {
		panic(rmdirErr)
	}
}

func TestLocationDoesntExist(t *testing.T) {
	removeRepoDir()
	logger := zerolog.New(os.Stdout)
	gitRepo := Git{location: localTestConfig.Location, logger: &logger}
	hasRepo := gitRepo.LocationHasRepo()
	if hasRepo {
		t.Errorf("gitRepo.locationHasRepo() = %v; want false", hasRepo)
	}
}
