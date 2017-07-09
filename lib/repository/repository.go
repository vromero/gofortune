package repository

import (
	"io/ioutil"
	"github.com/Masterminds/vcs"
	"os"
	"fmt"
	"path/filepath"
	"strings"
	"net/url"
	"regexp"
)

func InstallRepository(remote string) (err error) {
	repo, localRepoPath, _ := cloneRepositoryIntoTemp(remote)
	if err != nil {
		return err
	}

	installLocalRepository(repo, localRepoPath)
	if err != nil {
		return err
	}

	return nil
}

func installLocalRepository(repo vcs.Repo, localRepoPath string) (err error) {
	err = filepath.Walk(localRepoPath, func(path string, info os.FileInfo, err error) error {
		return visitRepositoryFile(repo, localRepoPath, path, info, err)
	})

	if err != nil {
		fmt.Println(err)
	}
	return
}


func visitRepositoryFile(repo vcs.Repo, basePath string, path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err)
		return nil
	}
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

	fmt.Println(getLocalFileName(repo, filepath.Base(relPath)))
	return nil
}

//func getLocalDirectoryName(repo vcs.Repo) {
//	localRepositoryDir, _ := ioutil.TempDir("", "gofortune")
//
//
//
//}

func getLocalFileName(repo vcs.Repo, fileName string) string {
	parserdUrl, err := url.Parse(repo.Remote())
	if err != nil {
		panic(err)
	}

	acceptableHostName, err := removeNonAcceptableChars(parserdUrl.Host)
	if err != nil {
		panic(err)
	}

	acceptablePath, err := removeNonAcceptableChars(parserdUrl.Path)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s%s_%s", acceptableHostName, acceptablePath, fileName)
}

func removeNonAcceptableChars(input string) (string, error) {
	reg, err := regexp.Compile("[^a-zA-Z0-9_]+")
	if err != nil {
		return "", err
	}
	return reg.ReplaceAllString(input, "_"), nil
}

func cloneRepositoryIntoTemp(remote string) (vcs.Repo, string, error) {

	tmpLocalRepo, _ := ioutil.TempDir("", "gofortune")

	repo, err := vcs.NewRepo(remote, tmpLocalRepo)
	if err != nil {
		panic(err)
	}

	repo.Get()
	if err != nil {
		panic(err)
	}
	return repo, tmpLocalRepo, nil
}