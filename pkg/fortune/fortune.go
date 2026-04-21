// Package fortune provides fortune cookie selection. This package will not output
// any data to the terminal.
package fortune

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/patrickdappollonio/localized"
	"github.com/vromero/gofortune/pkg"
)

type FortuneData struct {
	Data     string
	FileName string
}

type FortuneRequest struct {
	AllMaxims, ShowCookieFile, PrintListOfFiles bool
	LongDictumsOnly, ShortOnly, IgnoreCase      bool
	Wait, ConsiderAllEqual, Offensive           bool
	Match                                       string
	LongestShort                                int
	Paths                                       []ProbabilityPath
	OffensivePaths                              []ProbabilityPath
}

func PrepareRequest(args []string, defaultFortunePath, defaultOffensiveFortunePath string) FortuneRequest {
	request := FortuneRequest{}
	if len(args) == 0 {
		lang := localized.New()
		_ = lang.Detect()
		var lp, op string
		lp = filepath.Join(defaultFortunePath, lang.Lang)
		op = filepath.Join(defaultOffensiveFortunePath, lang.Lang)

		request.Paths = []ProbabilityPath{{Path: selectExisting(defaultFortunePath, lp)}}
		request.OffensivePaths = []ProbabilityPath{{Path: selectExisting(defaultOffensiveFortunePath, op)}}
	} else {
		currentPath := ProbabilityPath{}
		for i := range args {
			if strings.HasSuffix(args[i], "%") {
				value, err := strconv.ParseInt(strings.TrimSuffix(args[i], "%"), 10, 64)
				if err == nil {
					currentPath.Percentage = float32(value)
				}
			} else {
				currentPath.Path = args[i]
				request.Paths = append(request.Paths, currentPath)
				currentPath = ProbabilityPath{}
			}
		}
	}
	return request
}

func selectExisting(path1 string, path2 string) string {
	if path2 == "" || !pkg.FileExists(path2) {
		return path1
	}
	return path2
}

func GetRandomFortune(rootNode FileSystemNodeDescriptor) (FortuneData, error) {
	randomNode, err := GetRandomLeafNode(rootNode)
	if err != nil {
		return FortuneData{}, err
	}
	if randomNode.NumEntries == 0 {
		return FortuneData{}, errors.New("file is empty")
	}
	randomEntry := rand.Intn(int(randomNode.NumEntries))

	indexFile, err := os.Open(randomNode.IndexPath)
	if err != nil {
		return FortuneData{}, errors.New("can't open index file")
	}
	defer indexFile.Close()

	dataPos, err := pkg.ReadDataPos(indexFile, int(pkg.DataTableSize), uint32(randomEntry))
	if err != nil {
		return FortuneData{}, errors.New("can't read index file")
	}

	fortuneFile, err := os.Open(randomNode.Path)
	if err != nil {
		return FortuneData{}, errors.New("can't read fortune file")
	}
	defer fortuneFile.Close()

	fortuneData, err := pkg.ReadData(fortuneFile, int64(dataPos.OriginalOffset))
	if err != nil {
		return FortuneData{}, err
	}

	return FortuneData{FileName: filepath.Base(randomNode.Path), Data: fortuneData}, nil
}

func GetLengthFilteredRandomFortune(rootNode FileSystemNodeDescriptor, shorterThan uint32, longerThan uint32) (FortuneData, error) {

	for {
		fortune, err := GetRandomFortune(rootNode)
		if err != nil {
			return FortuneData{}, err
		}
		length := uint32(len(fortune.Data))
		if length > longerThan && length < shorterThan {
			return fortune, nil
		}
	}
}

func GetFortunesMatching(fsDescriptor FileSystemNodeDescriptor, expression string, ignoreCase bool) (<-chan FortuneData, <-chan error) {
	if ignoreCase {
		expression = "(?i)" + expression
	}
	matchingExpression := regexp.MustCompile(expression)
	return getFortunesMatching(fsDescriptor, matchingExpression)
}

func getFortunesMatching(fsDescriptor FileSystemNodeDescriptor, expression *regexp.Regexp) (<-chan FortuneData, <-chan error) {
	output := make(chan FortuneData)
	errorOutput := make(chan error)

	go func() {
		defer close(output)
		defer close(errorOutput)

		if len(fsDescriptor.Children) > 0 {
			for i := range fsDescriptor.Children {

				matchedDataChannel, matchedErrorChannel := getFortunesMatching(fsDescriptor.Children[i], expression)

				for result := range matchedDataChannel {
					output <- result
				}

				for errorResult := range matchedErrorChannel {
					errorOutput <- errorResult
				}
			}
			return
		}

		indexFile, err := os.Open(fsDescriptor.IndexPath)
		if err != nil {
			errorOutput <- errors.New("can't open index file")
			return
		}
		defer indexFile.Close()

		fortuneFile, err := os.Open(fsDescriptor.Path)
		if err != nil {
			errorOutput <- errors.New("can't read fortune file")
			return
		}
		defer fortuneFile.Close()

		for i := int64(0); i < int64(fsDescriptor.NumEntries); i++ {
			dataPos, err := pkg.ReadDataPos(indexFile, int(pkg.DataTableSize), uint32(i))
			if err != nil {
				errorOutput <- fmt.Errorf("can't read from index file, fortune number : %v", i)
				continue
			}

			data, err := pkg.ReadData(fortuneFile, int64(dataPos.OriginalOffset))
			if err != nil {
				errorOutput <- fmt.Errorf("can't read from fortune file, fortune number : %v", i)
				continue
			}
			if expression.MatchString(data) {
				output <- FortuneData{FileName: filepath.Base(fsDescriptor.Path), Data: data}
			}
		}
	}()

	return output, errorOutput
}
