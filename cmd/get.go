package cmd

import (
	"github.com/spf13/cobra"

	"fmt"
	"path/filepath"
	"os"
	"strings"
	"github.com/gofortune/gofortune/lib/repository"
)

var getName string = "get"
var getShortDescription string = "Downloads and installs a fortune cookie collection"
var getLongDescription string = `When fortune is run with no arguments it prints out a random epigram`

type GetRequest struct {
	RepoUrl string
}

var getRequest = GetRequest{}

// fortuneCmd represents the fortune command
var getCmd = &cobra.Command{
	Use:   getName,
	Short: getShortDescription,
	Long:  getLongDescription,
	Run: func(cmd *cobra.Command, args []string) {
 		getPrepareRequest(args)
		getRun(getRequest)
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
}

func getPrepareRequest(args []string) {
	getRequest.RepoUrl = args[0]
}

// fortuneRun executes fortune cookie operation requested in a FortuneRequest instance
func getRun(request GetRequest) (err error) {
	return repository.InstallRepository(request.RepoUrl)
}

func visitFile(basePath string, path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err)
		return nil
	}
	relPath, err := filepath.Rel(basePath, path)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if strings.HasPrefix(relPath, ".") && relPath != "." {
		return filepath.SkipDir
	}

	fmt.Println(relPath)
	return nil
}
