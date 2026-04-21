package repository

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/vcs"
)

func InstallRepository(remote string) error {
	repo, localRepoPath, err := cloneRepositoryIntoTemp(remote)
	if err != nil {
		return err
	}

	err = installLocalRepository(repo, localRepoPath)
	if err != nil {
		return err
	}

	return nil
}

func installLocalRepository(repo vcs.Repo, localRepoPath string) error {
	err := filepath.Walk(localFpath(localRepoPath), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return visitRepositoryFile(repo, localRepoPath, path, info)
	})

	return err
}

func localFpath(p string) string {
	return p
}

func visitRepositoryFile(repo vcs.Repo, basePath string, path string, info os.FileInfo) error {
	relPath, err := filepath.Rel(basePath, path)
	if err != nil {
		return err
	}

	if strings.HasPrefix(relPath, ".") && relPath != "." {
		return filepath.SkipDir
	}

	if info.IsDir() {
		return nil
	}

	name, err := getLocalFileName(repo, filepath.Base(relPath))
	if err != nil {
		return err
	}
	fmt.Println(name)
	return nil
}

func getLocalFileName(repo vcs.Repo, fileName string) (string, error) {
	parsedUrl, err := url.Parse(repo.Remote())
	if err != nil {
		return "", err
	}

	acceptableHostName, err := removeNonAcceptableChars(parsedUrl.Host)
	if err != nil {
		return "", err
	}

	acceptablePath, err := removeNonAcceptableChars(parsedUrl.Path)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s_%s", acceptableHostName, acceptablePath, fileName), nil
}

func removeNonAcceptableChars(input string) (string, error) {
	reg, err := regexp.Compile("[^a-zA-Z0-9_]+")
	if err != nil {
		return "", err
	}
	return reg.ReplaceAllString(input, "_"), nil
}

func cloneRepositoryIntoTemp(remote string) (vcs.Repo, string, error) {
	tmpLocalRepo, err := os.MkdirTemp("", "gofortune")
	if err != nil {
		return nil, "", err
	}

	repo, err := vcs.NewRepo(remote, tmpLocalRepo)
	if err != nil {
		_ = os.RemoveAll(tmpLocalRepo)
		return nil, "", err
	}

	err = repo.Get()
	if err != nil {
		_ = os.RemoveAll(tmpLocalRepo)
		return nil, "", err
	}
	return repo, tmpLocalRepo, nil
}
