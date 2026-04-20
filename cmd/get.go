package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vromero/gofortune/pkg/repository"
)

var getName = "get"
var getShortDescription = "Downloads and installs a fortune cookie collection"
var getLongDescription = `When fortune is run with no arguments it prints out a random epigram`

type GetRequest struct {
	RepoUrl string
}

var getRequest = GetRequest{}

// fortuneCmd represents the fortune command
var getCmd = &cobra.Command{
	Use:   getName,
	Short: getShortDescription,
	Long:  getLongDescription,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		getRequest.RepoUrl = args[0]
		return getRun(getRequest)
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
}


// fortuneRun executes fortune cookie operation requested in a FortuneRequest instance
func getRun(request GetRequest) (err error) {
	return repository.InstallRepository(request.RepoUrl)
}
