package cmd

import (
	"fmt"
	"os"

	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/patrickdappollonio/localized"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vromero/gofortune/pkg"
	"github.com/vromero/gofortune/pkg/fortune"
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

const (
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
		if err := fortuneRun(fortuneRequest); err != nil {
			panic(err)
		}
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
		var localizedDefaultFortunePath string
		var localizedDefaultOffensiveFortunePath string
		if errLangDetection == nil {
			localizedDefaultFortunePath = filepath.Join(defaultFortunePath, lang.Lang)
			localizedDefaultOffensiveFortunePath = filepath.Join(defaultOffensiveFortunePath, lang.Lang)
		}

		fortuneRequest.Paths = []fortune.ProbabilityPath{{Path: selectExisting(defaultFortunePath, localizedDefaultFortunePath)}}
		fortuneRequest.OffensivePaths = []fortune.ProbabilityPath{{Path: selectExisting(defaultOffensiveFortunePath, localizedDefaultOffensiveFortunePath)}}
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

// selectExisting selects the first existing path of the two paths passed
func selectExisting(path1 string, path2 string) string {
	if path2 == "" || !pkg.FileExists(path2) {
		return path1
	}
	return path2
}

// fortuneRun executes fortune cookie operation requested in a FortuneRequest instance
func fortuneRun(request FortuneRequest) (err error) {
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
		shorterThan uint32 = math.MaxUint32
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

	// Print out a random fortune from all the fortunes present in the directories and files
	// of the rootFsDescriptor graph. It will honor the possibilities data present in the graph.
	output, errorOutput := fortune.GetLengthFilteredRandomFortune(rootFsDescriptor, shorterThan, longerThan)
	printFortune(request, output, errorOutput)
	return nil
}

func printFortuneChannels(request FortuneRequest, fortuneChannel <-chan fortune.FortuneData, errorChannel <-chan error) {
	for fortuneData := range fortuneChannel {
		printFortune(request, fortuneData, nil)
	}

	for errorData := range errorChannel {
		printFortune(request, fortune.FortuneData{}, errorData)
	}
}

func printFortune(request FortuneRequest, fortune fortune.FortuneData, err error) {
	if err != nil {
		panic(err)
	}

	if request.ShowCookieFile {
		fmt.Printf("(%s)\n%%\n", fortune.FileName)
	}
	fmt.Println(fortune.Data)
	if request.Wait {
		readTimeWait(len(fortune.Data))
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
// The minimum time wait is defined as constant in this file
func readTimeWait(length int) {
	timeWait := pkg.Max(uint32(length/charsPerSec), uint32(minimumWaitSeconds))
	time.Sleep(time.Second * time.Duration(timeWait))
}
