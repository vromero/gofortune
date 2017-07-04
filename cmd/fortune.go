package cmd

import (
	"fmt"

	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"strconv"

	"github.com/gofortune/gofortune/lib"
	"github.com/gofortune/gofortune/lib/fortune"
)

type FortuneRequest struct {
	AllMaxims, ShowCookieFile, PrintListOfFiles bool
	LongDictumsOnly, ShortOnly, IgnoreCase      bool
	Wait, ConsiderAllEqual, Offensive           bool
	Match                                       string
	LongestShort                                int
	Paths                                       []fortune.ProbabilityPath
	OffensivePaths                              []fortune.ProbabilityPath
}

var fortuneRequest = FortuneRequest{}

var (
	defaultFortunePath          = "/usr/share/games/fortunes"
	defaultOffensiveFortunePath = "/usr/share/games/fortunes/off"
	minimumWaitSeconds          = 6
	charsPerSec                 = 20
)

var fortuneName string = "fortune"
var fortuneShortDescription string = "Print a random, hopefully interesting, adage"
var fortuneLongDescription string = `When fortune is run with no arguments it prints out a random epigram`

// fortuneCmd represents the fortune command
var fortuneCmd = &cobra.Command{
	Use:   fortuneName,
	Short: fortuneShortDescription,
	Long:  fortuneLongDescription,
	Run: func(cmd *cobra.Command, args []string) {
		fortunePrepareRequest(args)
		fortuneRun(fortuneRequest)
	},
}

func init() {
	RootCmd.AddCommand(fortuneCmd)
	fortuneCmd.Flags().BoolVarP(&fortuneRequest.AllMaxims, "allMaxims", "a", false, "Choose from all lists of maxims")
	fortuneCmd.Flags().BoolVarP(&fortuneRequest.Offensive, "offensive", "o", false, "Choose only from potentially offensive aphorisms")
	fortuneCmd.Flags().BoolVarP(&fortuneRequest.ShowCookieFile, "showCookieFile", "c", false, "Show the cookie file from which the fortune came")
	fortuneCmd.Flags().BoolVarP(&fortuneRequest.PrintListOfFiles, "printListOfFiles", "f", false, "Print out the list of files which would be searched, but don't print a fortune")
	fortuneCmd.Flags().BoolVarP(&fortuneRequest.ConsiderAllEqual, "considerAllEqual", "e", false, "Consider all fortune files to be of equal size")
	fortuneCmd.Flags().StringVarP(&fortuneRequest.Match, "match", "m", "", "Print out all fortunes which match the basic regular expression pattern")
	fortuneCmd.Flags().IntVarP(&fortuneRequest.LongestShort, "longestShort", "n", 160, "set the longest fortune length (in characters) considered to be \"short\" (the default is 160)")
	fortuneCmd.Flags().BoolVarP(&fortuneRequest.LongDictumsOnly, "longDictumsOnly", "l", false, "Long dictums only. See -n on how \"long\" is defined in this sense")
	fortuneCmd.Flags().BoolVarP(&fortuneRequest.ShortOnly, "shortOnly", "s", false, "Short apothegms only. See -n on which fortunes are considered \"short\"")
	fortuneCmd.Flags().BoolVarP(&fortuneRequest.IgnoreCase, "ignoreCase", "i", false, "Ignore case for -m patterns")
	fortuneCmd.Flags().BoolVarP(&fortuneRequest.Wait, "wait", "w", false, "Wait before termination for an amount of time calculated from the number of characters in the message")
}

func fortunePrepareRequest(args []string) {
	if len(args) == 0 {
		fortuneRequest.Paths = []fortune.ProbabilityPath{{Path: defaultFortunePath}}
		fortuneRequest.OffensivePaths = []fortune.ProbabilityPath{{Path: defaultOffensiveFortunePath}}
	} else {
		currentPath := fortune.ProbabilityPath{}
		for i := range args {
			if strings.HasSuffix(args[i], "%") {
				value, err := strconv.ParseInt(strings.TrimSuffix(args[i], "%"), 10, 64)
				if err == nil {
					currentPath.Percentage = float32(value)
				}
			} else {
				currentPath.Path = args[i]
				fortuneRequest.Paths = append(fortuneRequest.Paths, currentPath)
				currentPath = fortune.ProbabilityPath{}
			}
		}
	}
}

// fortuneRun executes fortune cookie operation requested in a FortuneRequest instance
func fortuneRun(request FortuneRequest) (err error) {
	input := []fortune.ProbabilityPath{}

	if request.AllMaxims {
		input = append(input, request.Paths...)
		input = append(input, request.OffensivePaths...)
	} else if request.Offensive {
		input = append(input, request.OffensivePaths...)
	} else {
		input = append(input, request.Paths...)
	}

	rootFsDescriptor, err := fortune.LoadPaths(input)

	if request.Match != "" {
		fortune.MatchFortunes(rootFsDescriptor, request.Match, func(input string) {
			fmt.Println(input)
			fmt.Println("%")
		})
		os.Exit(0)
	}

	fortune.SetProbabilities(&rootFsDescriptor, request.ConsiderAllEqual)

	if request.PrintListOfFiles {
		printListOfFiles(rootFsDescriptor)
		os.Exit(0)
	}

	length := printRandomFortune(request, rootFsDescriptor)

	if request.Wait {
		readTimeWait(length)
	}

	return nil
}

// Print out a random fortune from all the fortunes present in the directories and files
// of the rootFsDescriptor graph. It will honor the possibilities data present in the graph.
func printRandomFortune(request FortuneRequest, rootFsDescriptor fortune.FileSystemNodeDescriptor) int {
	var fortuneData, fileName string
	filter := func(f string) bool {
		return (!request.ShortOnly && !request.LongDictumsOnly) ||
			(request.ShortOnly && len(f) < request.LongestShort ||
				request.LongDictumsOnly && len(f) > request.LongestShort)
	}
	fileName, fortuneData, _ = fortune.GetFilteredRandomFortune(rootFsDescriptor, filter)

	if request.ShowCookieFile {
		fmt.Printf("(%s)\n%%\n", fileName)
	}

	fmt.Println(fortuneData)
	return len(fortuneData)
}

// Wait a length relative amount of time after fortune is printed.
// The minimum time wait is defined
func readTimeWait(length int) {
	timeWait := lib.Max(uint32(length/charsPerSec), uint32(minimumWaitSeconds))
	time.Sleep(time.Second * time.Duration(timeWait))
}

// Print out the list of directories and files, including its possibilities data.
func printListOfFiles(directoryDescriptor fortune.FileSystemNodeDescriptor) {
	for i := range directoryDescriptor.Children {
		fmt.Printf("%5.2f%% %s\n", directoryDescriptor.Children[i].Percent, directoryDescriptor.Children[i].Path)
		for j := range directoryDescriptor.Children[i].Children {
			fmt.Printf("%*s", 4, "")
			fmt.Printf("%5.2f%% %s\n", directoryDescriptor.Children[i].Children[j].Percent, filepath.Base(directoryDescriptor.Children[i].Children[j].Path))
		}
	}
}
