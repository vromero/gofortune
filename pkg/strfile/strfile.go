// Package strfile provides index file generation logic. This package will not output
// any data to the terminal.
package strfile

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/vromero/gofortune/pkg"
)

func StrFile(ignoreCase bool, silent bool, order bool, randomize bool, rot13 bool, delimitingChar string, sourceFile string, dataFile string) (err error) {
	inputFile, err := os.Open(sourceFile)
	defer inputFile.Close()

	outputFile, err := os.Create(dataFile)
	defer outputFile.Close()

	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(inputFile)
	scanner.Split(advanceAwareSplitter)

	var totalFortunes, longestFortune uint32
	var shortestFortune uint32 = math.MaxUint32
	var pos uint32
	var fortuneBytes []byte

	fortuneBase := make([]pkg.DataPos, 0)

	for scanner.Scan() {
		fortunePortion := scanner.Bytes()
		pos += uint32(len(fortunePortion))

		if string(pkg.RemoveCRLF(fortunePortion)) == delimitingChar {
			totalFortunes++
			fortuneStringLength := uint32(len(fortuneBytes))
			shortestFortune = pkg.Min(shortestFortune, fortuneStringLength)
			longestFortune = pkg.Max(longestFortune, fortuneStringLength)

			transformedString := applyFortuneTransformations(string(fortuneBytes), ignoreCase, rot13)
			dataPos := pkg.DataPos{OriginalOffset: pos, Text: transformedString}
			if !order && !randomize {
				pkg.WriteDataPos(outputFile, int(pkg.DataTableSize), totalFortunes, dataPos)
			} else {
				fortuneBase = append(fortuneBase, dataPos)
			}

			fortuneBytes = make([]byte, 0)
		} else {
			fortuneBytes = append(fortuneBytes[:], fortunePortion[:]...)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	flags := calculateFlags(randomize, order, rot13)
	posContents := pkg.CreateDataTable(totalFortunes, longestFortune, shortestFortune, flags, delimitingChar)
	pkg.SaveDataTable(outputFile, posContents)

	if !silent {
		report(dataFile, totalFortunes, longestFortune, shortestFortune)
	}

	if order {
		sort.Slice(fortuneBase, func(i, j int) bool {
			return pkg.LessThanDataPos(fortuneBase[i], fortuneBase[j])
		})
	} else if randomize {
		Shuffle(fortuneBase)
	}

	if order || randomize {
		pkg.WriteDataPosSlice(outputFile, int(pkg.DataTableSize), fortuneBase)
	}

	return
}

func calculateFlags(randomize bool, order bool, rot13 bool) (flags uint32) {
	if randomize {
		flags = flags | pkg.FlagRandom
	}

	if order {
		flags = flags | pkg.FlagOrdered
	}

	if rot13 {
		flags = flags | pkg.FlagRotated
	}
	return
}

func applyFortuneTransformations(input string, ignoreCase bool, rot13 bool) (output string) {
	output = input
	if ignoreCase {
		output = strings.ToLower(output)
	}

	if rot13 {
		output = pkg.Rot13(input)
	}
	return
}

// report prints the result summary of the strfile operation.
func report(fileName string, totalFortunes uint32, longestFortune uint32, shortestFortune uint32) {
	fmt.Printf("\"%s\" created\n", fileName)

	switch totalFortunes {
	case 0:
		fmt.Println("There was no string")
		return
	case 1:
		fmt.Println("There was 1 string")
	default:
		fmt.Printf("There were %d strings\n", totalFortunes)
	}

	fmt.Printf("Longest string: %d bytes\n", longestFortune)
	fmt.Printf("Shortest string: %d bytes\n", shortestFortune)
}
