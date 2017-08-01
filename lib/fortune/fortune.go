// Package fortune provides fortune cookie selection. This package will not output
// any data to the terminal.
package fortune

import (
	"math/rand"
	"os"

	"path/filepath"
	"time"

	"regexp"

	"fmt"

	"github.com/gofortune/gofortune/lib"
)

type FortuneData struct {
	Data       string
	FileName   string
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func GetRandomFortune(rootNode FileSystemNodeDescriptor) <-chan FortuneData {
	output := make(chan FortuneData)

	go func() {
		randomNode := GetRandomLeafNode(rootNode)
		if randomNode.NumEntries == 0 {
			panic("File is empty")
		}
		randomEntry := rand.Intn(int(randomNode.NumEntries))

		indexFile, err := os.Open(randomNode.IndexPath)
		defer indexFile.Close()
		if err != nil {
			panic("Can't open index file")
		}

		dataPos, err := lib.ReadDataPos(indexFile, int(lib.DataTableSize), uint32(randomEntry))
		if err != nil {
			panic("Can't read index file")
		}
		indexFile.Close()

		fortuneFile, err := os.Open(randomNode.Path)
		if err != nil {
			panic("Can't read fortune file")
		}
		defer fortuneFile.Close()

		fortuneData, err := lib.ReadData(fortuneFile, int64(dataPos.OriginalOffset))

		output <- FortuneData{FileName: filepath.Base(randomNode.Path),Data: fortuneData}
		close(output)
	}()

	return output
}

func GetLengthFilteredRandomFortune(rootNode FileSystemNodeDescriptor, shorterThan uint32, longerThan uint32) <-chan FortuneData {
	output := make(chan FortuneData)
	go func() {
		for {
			n := <- GetRandomFortune(rootNode)
			length := uint32(len(n.Data))
			if length > longerThan && length < shorterThan {
				output <- n
				break
			}
		}
		close(output)
	}()
	return output
}

func MatchFortunes(fsDescriptor FileSystemNodeDescriptor, expression string, ignoreCase bool) <-chan FortuneData {
	if ignoreCase {
		expression = "(?i)" + expression
	}
	matchingExpression := regexp.MustCompile(expression)
	return matchFortunesWithCompiledExpression(fsDescriptor, matchingExpression)
}

func matchFortunesWithCompiledExpression(fsDescriptor FileSystemNodeDescriptor, expression *regexp.Regexp) <-chan FortuneData {
	output := make(chan FortuneData)

	go func() {
		if len(fsDescriptor.Children) > 0 {
			for i := range fsDescriptor.Children {
				for n := range matchFortunesWithCompiledExpression(fsDescriptor.Children[i], expression) {
					output <- n
				}
			}
		} else {
			indexFile, err := os.Open(fsDescriptor.IndexPath)
			defer indexFile.Close()
			if err != nil {
				panic("Can't open index file")
			}

			fortuneFile, err := os.Open(fsDescriptor.Path)
			defer fortuneFile.Close()
			if err != nil {
				panic("Can't read fortune file")
			}

			for i := int64(0); i < int64(fsDescriptor.NumEntries); i++ {
				dataPos, err := lib.ReadDataPos(indexFile, int(lib.DataTableSize), uint32(i))
				if err != nil {
					panic(fmt.Sprintf("Can't read from index file, fortune number : %v", i))
				}

				data, err := lib.ReadData(fortuneFile, int64(dataPos.OriginalOffset))
				if expression.MatchString(data) {
					output <- FortuneData{FileName: filepath.Base(fsDescriptor.Path), Data: data}
				}
			}
		}
		close(output)
	}()

	return output
}
