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
	"time"

	"github.com/vromero/gofortune/pkg"
)

type FortuneData struct {
	Data     string
	FileName string
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func GetRandomFortune(rootNode FileSystemNodeDescriptor) (FortuneData, error) {
	randomNode := GetRandomLeafNode(rootNode)
	if randomNode.NumEntries == 0 {
		return FortuneData{}, errors.New("file is empty")
	}
	randomEntry := rand.Intn(int(randomNode.NumEntries))

	indexFile, err := os.Open(randomNode.IndexPath)
	defer indexFile.Close()
	if err != nil {
		return FortuneData{}, errors.New("can't open index file")
	}

	dataPos, err := pkg.ReadDataPos(indexFile, int(pkg.DataTableSize), uint32(randomEntry))
	if err != nil {
		return FortuneData{}, errors.New("can't read index file")
	}
	indexFile.Close()

	fortuneFile, err := os.Open(randomNode.Path)
	if err != nil {
		return FortuneData{}, errors.New("can't read fortune file")
	}
	defer fortuneFile.Close()

	fortuneData, err := pkg.ReadData(fortuneFile, int64(dataPos.OriginalOffset))

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
		} else {
			indexFile, err := os.Open(fsDescriptor.IndexPath)
			defer indexFile.Close()
			if err != nil {
				errorOutput <- errors.New("can't open index file")
			}

			fortuneFile, err := os.Open(fsDescriptor.Path)
			defer fortuneFile.Close()
			if err != nil {
				errorOutput <- errors.New("can't read fortune file")
			}

			for i := int64(0); i < int64(fsDescriptor.NumEntries); i++ {
				dataPos, err := pkg.ReadDataPos(indexFile, int(pkg.DataTableSize), uint32(i))
				if err != nil {
					errorOutput <- errors.New(fmt.Sprintf("can't read from index file, fortune number : %v", i))
				}

				data, err := pkg.ReadData(fortuneFile, int64(dataPos.OriginalOffset))
				if expression.MatchString(data) {
					output <- FortuneData{FileName: filepath.Base(fsDescriptor.Path), Data: data}
				}
			}
		}
		close(output)
		close(errorOutput)
	}()

	return output, errorOutput
}
