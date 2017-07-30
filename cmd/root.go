package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/gofortune/gofortune/lib/fortune"
	"time"
	"path/filepath"
	"strings"

	"github.com/gofortune/gofortune/lib"
	"strconv"
	"github.com/patrickdappollonio/localized"
)

var cfgFile string

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


var RootCmd = &cobra.Command{
	Use:   "gofortune",
	Short: "Print a random, hopefully interesting, adage",
	Long:  `When fortune is run with no arguments it prints out a random epigram`,
	Run: func(cmd *cobra.Command, args []string) {
		fortunePrepareRequest(args)
		fortuneRun(fortuneRequest)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.Flags().BoolVarP(&fortuneRequest.AllMaxims, "allMaxims", "a", false, "Choose from all lists of maxims")
	RootCmd.Flags().BoolVarP(&fortuneRequest.Offensive, "offensive", "o", false, "Choose only from potentially offensive aphorisms")
	RootCmd.Flags().BoolVarP(&fortuneRequest.ShowCookieFile, "showCookieFile", "c", false, "Show the cookie file from which the fortune came")
	RootCmd.Flags().BoolVarP(&fortuneRequest.PrintListOfFiles, "printListOfFiles", "f", false, "Print out the list of files which would be searched, but don't print a fortune")
	RootCmd.Flags().BoolVarP(&fortuneRequest.ConsiderAllEqual, "considerAllEqual", "e", false, "Consider all fortune files to be of equal size")
	RootCmd.Flags().StringVarP(&fortuneRequest.Match, "match", "m", "", "Print out all fortunes which match the basic regular expression pattern")
	RootCmd.Flags().IntVarP(&fortuneRequest.LongestShort, "longestShort", "n", 160, "set the longest fortune length (in characters) considered to be \"short\" (the default is 160)")
	RootCmd.Flags().BoolVarP(&fortuneRequest.LongDictumsOnly, "longDictumsOnly", "l", false, "Long dictums only. See -n on how \"long\" is defined in this sense")
	RootCmd.Flags().BoolVarP(&fortuneRequest.ShortOnly, "shortOnly", "s", false, "Short apothegms only. See -n on which fortunes are considered \"short\"")
	RootCmd.Flags().BoolVarP(&fortuneRequest.IgnoreCase, "ignoreCase", "i", false, "Ignore case for -m patterns")
	RootCmd.Flags().BoolVarP(&fortuneRequest.Wait, "wait", "w", false, "Wait before termination for an amount of time calculated from the number of characters in the message")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with fortuneName ".gofortune" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".gofortune")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// fortunePrepareRequest setups fortuneRequest.Paths and fortuneRequest.OffensivePaths taking
// into consideration user input and environmental language information
func fortunePrepareRequest(args []string) {
	if len(args) == 0 {
		// If no arguments are passed, localizedDefaultFortunePath / environment language will be tried, if they
		// don't exist the regular ones will be tried
		lang := localized.New()
		errLangDetection := lang.Detect()

		localizedDefaultFortunePath := filepath.Join(defaultFortunePath, lang.Lang)
		localizedDefaultOffensiveFortunePath := filepath.Join(defaultOffensiveFortunePath, lang.Lang)
		if errLangDetection == nil && (lib.FileExists(localizedDefaultFortunePath) ||
			lib.FileExists(localizedDefaultOffensiveFortunePath)) {
			fortuneRequest.Paths = []fortune.ProbabilityPath{{Path: localizedDefaultFortunePath}}
			fortuneRequest.OffensivePaths = []fortune.ProbabilityPath{{Path: localizedDefaultOffensiveFortunePath}}
		} else {
			fortuneRequest.Paths = []fortune.ProbabilityPath{{Path: defaultFortunePath}}
			fortuneRequest.OffensivePaths = []fortune.ProbabilityPath{{Path: defaultOffensiveFortunePath}}
		}
	} else {
		// If arguments are passed those will be used as directories. They may contain specific probabilities.
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
		matchedFortunes := fortune.MatchFortunes(rootFsDescriptor, request.Match, request.IgnoreCase)
		printFortuneChannel(request, matchedFortunes)
		os.Exit(0)
	}

	fortune.SetProbabilities(&rootFsDescriptor, request.ConsiderAllEqual)

	if request.PrintListOfFiles {
		printListOfFiles(rootFsDescriptor)
		os.Exit(0)
	}

	printRandomFortune(request, rootFsDescriptor)
	return nil
}

// Print out a random fortune from all the fortunes present in the directories and files
// of the rootFsDescriptor graph. It will honor the possibilities data present in the graph.
func printRandomFortune(request FortuneRequest, rootFsDescriptor fortune.FileSystemNodeDescriptor) {
	filter := func(f string) bool {
		return (!request.ShortOnly && !request.LongDictumsOnly) ||
			(request.ShortOnly && len(f) < request.LongestShort ||
				request.LongDictumsOnly && len(f) > request.LongestShort)
	}
	output := fortune.GetFilteredRandomFortune(rootFsDescriptor, filter)
	printFortuneChannel(request, output)
}

func printFortuneChannel(request FortuneRequest, fortuneChannel <- chan fortune.FortuneData) {
	for fortuneData := range fortuneChannel {
		if request.ShowCookieFile {
			fmt.Printf("(%s)\n%%\n", fortuneData.FileName)
		}
		fmt.Println(fortuneData.Data)
		if request.Wait {
			readTimeWait(len(fortuneData.Data))
		}
	}
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

// Wait a length relative amount of time after fortune is printed.
// The minimum time wait is defined
func readTimeWait(length int) {
	timeWait := lib.Max(uint32(length/charsPerSec), uint32(minimumWaitSeconds))
	time.Sleep(time.Second * time.Duration(timeWait))
}
