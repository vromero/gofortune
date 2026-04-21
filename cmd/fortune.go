package cmd

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
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

// rootFlags holds values bound to RootCmd's flags. Bound directly on the
// cobra flag set (StringVarP/BoolVarP) to avoid repetitive GetX/err-drop
// boilerplate in RunE.
var rootFlags struct {
	AllMaxims        bool
	Offensive        bool
	ShowCookieFile   bool
	PrintListOfFiles bool
	ConsiderAllEqual bool
	Match            string
	LongestShort     int
	LongDictumsOnly  bool
	ShortOnly        bool
	IgnoreCase       bool
	Wait             bool
}

var RootCmd = &cobra.Command{
	Use:   "gofortune",
	Short: "Print a random, hopefully interesting, adage",
	Long:  `When fortune is run with no arguments it prints out a random epigram`,
	RunE: func(cmd *cobra.Command, args []string) error {
		request, err := fortune.PrepareRequest(args, defaultFortunePath, defaultOffensiveFortunePath)
		if err != nil {
			return err
		}

		request.AllMaxims = rootFlags.AllMaxims
		request.Offensive = rootFlags.Offensive
		request.ShowCookieFile = rootFlags.ShowCookieFile
		request.PrintListOfFiles = rootFlags.PrintListOfFiles
		request.ConsiderAllEqual = rootFlags.ConsiderAllEqual
		request.Match = rootFlags.Match
		request.LongestShort = rootFlags.LongestShort
		request.LongDictumsOnly = rootFlags.LongDictumsOnly
		request.ShortOnly = rootFlags.ShortOnly
		request.IgnoreCase = rootFlags.IgnoreCase
		request.Wait = rootFlags.Wait

		return fortuneRun(request)
	},
}

func Execute() error {
	return RootCmd.Execute()
}

func init() {
	if runtime.GOOS == "windows" {
		configDir, err := os.UserConfigDir()
		if err == nil {
			defaultFortunePath = filepath.Join(configDir, "gofortune", "fortunes")
			defaultOffensiveFortunePath = filepath.Join(configDir, "gofortune", "fortunes", "off")
		} else {
			defaultFortunePath = `C:\ProgramData\gofortune\fortunes`
			defaultOffensiveFortunePath = `C:\ProgramData\gofortune\fortunes\off`
		}
	}

	f := RootCmd.Flags()
	f.BoolVarP(&rootFlags.AllMaxims, "allMaxims", "a", false, "Choose from all lists of maxims")
	f.BoolVarP(&rootFlags.Offensive, "offensive", "o", false, "Choose only from potentially offensive aphorisms")
	f.BoolVarP(&rootFlags.ShowCookieFile, "showCookieFile", "c", false, "Show the cookie file from which the fortune came")
	f.BoolVarP(&rootFlags.PrintListOfFiles, "printListOfFiles", "f", false, "Print out the list of files which would be searched, but don't print a fortune")
	f.BoolVarP(&rootFlags.ConsiderAllEqual, "considerAllEqual", "e", false, "Consider all fortune files to be of equal size")
	f.StringVarP(&rootFlags.Match, "match", "m", "", "Print out all fortunes which match the regular expression pattern")
	f.IntVarP(&rootFlags.LongestShort, "longestShort", "n", 160, "set the longest fortune length (in characters) considered to be \"short\" (the default is 160)")
	f.BoolVarP(&rootFlags.LongDictumsOnly, "longDictumsOnly", "l", false, "Long dictums only. See -n on how \"long\" is enough")
	f.BoolVarP(&rootFlags.ShortOnly, "shortOnly", "s", false, "Short apothegms only. See -n on which fortunes are considered \"short\"")
	f.BoolVarP(&rootFlags.IgnoreCase, "ignoreCase", "i", false, "Ignore case for -m patterns")
	f.BoolVarP(&rootFlags.Wait, "wait", "w", false, "Wait before termination for an amount of time calculated from the number of characters in the message")
}

func fortuneRun(request fortune.Request) error {
	var input []fortune.ProbabilityPath
	switch {
	case request.AllMaxims:
		input = append(input, request.Paths...)
		input = append(input, request.OffensivePaths...)
	case request.Offensive:
		input = append(input, request.OffensivePaths...)
	default:
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
		return nil
	}

	fortune.SetProbabilities(&rootFsDescriptor, request.ConsiderAllEqual)

	if request.PrintListOfFiles {
		printListOfFiles(rootFsDescriptor)
		return nil
	}

	output, err := fortune.GetLengthFilteredRandomFortune(rootFsDescriptor, shorterThan, longerThan)
	if err != nil {
		return err
	}
	printFortune(request, output, nil)
	return nil
}

// printFortuneChannels drains both the fortune and error channels
// concurrently using a single select loop. Draining them sequentially would
// risk deadlocking the producer if it blocks trying to send on a channel
// whose consumer hasn't started reading yet.
func printFortuneChannels(request fortune.Request, fortuneChannel <-chan fortune.Cookie, errorChannel <-chan error) {
	for fortuneChannel != nil || errorChannel != nil {
		select {
		case cookie, ok := <-fortuneChannel:
			if !ok {
				fortuneChannel = nil
				continue
			}
			printFortune(request, cookie, nil)
		case err, ok := <-errorChannel:
			if !ok {
				errorChannel = nil
				continue
			}
			printFortune(request, fortune.Cookie{}, err)
		}
	}
}

func printFortune(request fortune.Request, cookie fortune.Cookie, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if request.ShowCookieFile {
		fmt.Printf("(%s)\n%%\n", cookie.FileName)
	}
	fmt.Println(cookie.Data)
	if request.Wait {
		readTimeWait(len(cookie.Data))
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
