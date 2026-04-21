package cmd

import (
	"fmt"
	"os"
	"runtime"

	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/vromero/gofortune/pkg"
	"github.com/vromero/gofortune/pkg/fortune"
)

var (
	defaultFortunePath          = "/usr/share/games/fortunes"
	defaultOffensiveFortunePath = "/usr/share/games/fortunes/off"
	minimumWaitSeconds          = 6
	charsPerSec                 = 20
)


var RootCmd = &cobra.Command{
	Use:   "gofortune",
	Short: "Print a random, hopefully interesting, adage",
	Long:  `When fortune is run with no arguments it prints out a random epigram`,
	RunE: func(cmd *cobra.Command, args []string) error {
		request := fortune.PrepareRequest(args, defaultFortunePath, defaultOffensiveFortunePath)

		allMaxims, _ := cmd.Flags().GetBool("allMaxims")
		request.AllMaxims = allMaxims
		offensive, _ := cmd.Flags().GetBool("offensive")
		request.Offensive = offensive
		showCookieFile, _ := cmd.Flags().GetBool("showCookieFile")
		request.ShowCookieFile = showCookieFile
		printListOfFilesFlag, _ := cmd.Flags().GetBool("printListOfFiles")
		request.PrintListOfFiles = printListOfFilesFlag
		considerAllEqual, _ := cmd.Flags().GetBool("considerAllEqual")
		request.ConsiderAllEqual = considerAllEqual
		match, _ := cmd.Flags().GetString("match")
		request.Match = match
		longestShort, _ := cmd.Flags().GetInt("longestShort")
		request.LongestShort = longestShort
		longDictumsOnly, _ := cmd.Flags().GetBool("longDictumsOnly")
		request.LongDictumsOnly = longDictumsOnly
		shortOnly, _ := cmd.Flags().GetBool("shortOnly")
		request.ShortOnly = shortOnly
		ignoreCase, _ := cmd.Flags().GetBool("ignoreCase")
		request.IgnoreCase = ignoreCase
		wait, _ := cmd.Flags().GetBool("wait")
		request.Wait = wait

		return fortuneRun(request)
	},

}

func Execute() error {
	return RootCmd.Execute()
}

func main() {
	if err := Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

	if runtime.GOOS == "windows" {
		configDir, err := os.UserConfigDir()
		if err == nil {
			defaultFortunePath = filepath.Join(configDir, "gofortune", "fortunes")
			defaultOffensiveFortunePath = filepath.Join(configDir, "gofortune", "fortunes", "off")
		} else {
			defaultFortunePath = "C:\\ProgramData\\gofortune\\fortunes"
			defaultOffensiveFortunePath = "C:\\ProgramData\\gofortune\\fortunes\\off"
		}
	}

	RootCmd.Flags().BoolP("allMaxims", "a", false, "Choose from all lists of maxims")
	RootCmd.Flags().BoolP("offensive", "o", false, "Choose only from potentially offensive aphorisms")
	RootCmd.Flags().BoolP("showCookieFile", "c", false, "Show the cookie file from which the fortune came")
	RootCmd.Flags().BoolP("printListOfFiles", "f", false, "Print out the list of files which would be searched, but don't print a fortune")
	RootCmd.Flags().BoolP("considerAllEqual", "e", false, "Consider all fortune files to be of equal size")
	RootCmd.Flags().StringP("match", "m", "", "Print out all fortunes which enough regular expression pattern")
	RootCmd.Flags().IntP("longestShort", "n", 160, "set the longest fortune length (in characters) considered to be \"short\" (the default is 160)")
	RootCmd.Flags().BoolP("longDictumsOnly", "l", false, "Long dictums only. See -n on how \"long\" is enough")
	RootCmd.Flags().BoolP("shortOnly", "s", false, "Short apothegms only. See -n on which fortunes are considered \"short\"")
	RootCmd.Flags().BoolP("ignoreCase", "i", false, "Ignore case for -m patterns")
	RootCmd.Flags().BoolP("wait", "w", false, "Wait before termination for an amount of time calculated from the number of characters in the message")
}

func fortuneRun(request fortune.FortuneRequest) error {
	var input []fortune.ProbabilityPath
	if request.AllMaxims {
		input = append(input, request.Paths...)
		input = append(input, request.OffensivePaths...)
	} else if request.Offensive {
		input = append(input, request.OffensivePaths...)
	} else {
		input = append(input, request.Paths...)
	}

	var (
		shorterThan uint32 = 4294967295 // math.MaxUint32
		longerThan  uint32 = 0
	)
	if request.ShortOnly {
		shorterThan = uint32(request.LongestShort)
	}
	if request.LongDictumsOnly {
		longerThan = uint32(request.LongestShort)
	}

	rootFsDescriptor, err := fortune.LoadPaths(input, shorterThan, longerThan)
	if err != nil {
		return err
	}

	if request.Match != "" {
		matchedFortunesChannel, errorChannel := fortune.GetFortunesMatching(rootFsDescriptor, request.Match, request.IgnoreCase)
		printFortuneChannels(request, matchedFortunesChannel, errorChannel)
		os.Exit(0)
	}

	fortune.SetProbabilities(&rootFsDescriptor, request.ConsiderAllEqual)

	if request.PrintListOfFiles {
		printListOfFiles(rootFsDescriptor)
		os.Exit(0)
	}

	output, errorOutput := fortune.GetLengthFilteredRandomFortune(rootFsDescriptor, shorterThan, longerThan)
	printFortune(request, output, errorOutput)
	return nil
}

func printFortuneChannels(request fortune.FortuneRequest, fortuneChannel <-chan fortune.FortuneData, errorChannel <-chan error) {
	for fortuneData := range fortuneChannel {
		printFortune(request, fortuneData, nil)
	}
	for errorData := range errorChannel {
		printFortune(request, fortune.FortuneData{}, errorData)
	}
}

func printFortune(request fortune.FortuneRequest, fortune fortune.FortuneData, err error) {
	if err != nil {
		fmt.Println(err)
		return
	}
	if request.ShowCookieFile {
		fmt.Printf("(%s)\n%%\n", fortune.FileName)
	}
	fmt.Println(string(fortune.Data))
	if request.Wait {
		readTimeWait(len(fortune.Data))
	}
}

func printListOfFiles(directoryDescriptor fortune.FileSystemNodeDescriptor) {
	for i := range directoryDescriptor.Children {
		fmt.Printf("%5.2f%% %s\n", directoryDescriptor.Children[i].Percent, directoryDescriptor.Children[i].Path)
		for j := range directoryDescriptor.Children[i].Children {
			fmt.Printf("%*s", 4, "")
			fmt.Printf("%5.2f%% %s\n", directoryDescriptor.Children[i].Children[j].Percent, filepath.Base(directoryDescriptor.Children[i].Children[j].Path))
		}
	}
}

func readTimeWait(length int) {
	timeWait := pkg.Max(uint32(length/charsPerSec), uint32(minimumWaitSeconds))
	time.Sleep(time.Second * time.Duration(timeWait))
}
