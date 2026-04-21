// Package strfile provides index file generation logic. This package will not output
// any data to the terminal; callers receive a Summary describing the result
// and decide whether (and how) to present it to the user.
package strfile

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/vromero/gofortune/pkg"
)

// Summary describes the result of building an index from a fortune file.
type Summary struct {
	DataFile        string
	TotalFortunes   uint32
	LongestFortune  uint32
	ShortestFortune uint32
}

// WriteTo writes a human-readable report of s to w, mimicking the classic
// strfile(1) output. It returns the number of bytes written and any error
// encountered, so callers can chain it behind a silent flag if desired.
func (s Summary) WriteTo(w io.Writer) (int64, error) {
	var total int64
	write := func(format string, args ...any) error {
		n, err := fmt.Fprintf(w, format, args...)
		total += int64(n)
		return err
	}

	if err := write("%q created\n", s.DataFile); err != nil {
		return total, err
	}
	switch s.TotalFortunes {
	case 0:
		if err := write("There was no string\n"); err != nil {
			return total, err
		}
		return total, nil
	case 1:
		if err := write("There was 1 string\n"); err != nil {
			return total, err
		}
	default:
		if err := write("There were %d strings\n", s.TotalFortunes); err != nil {
			return total, err
		}
	}
	if err := write("Longest string: %d bytes\n", s.LongestFortune); err != nil {
		return total, err
	}
	if err := write("Shortest string: %d bytes\n", s.ShortestFortune); err != nil {
		return total, err
	}
	return total, nil
}

// StrFile builds a fortune index at dataFile from sourceFile, returning a
// Summary describing the index. The silent parameter has been removed from
// this function signature; see cmd/strfile.go for user-facing silence
// handling.
func StrFile(ignoreCase bool, order bool, randomize bool, rot13 bool, delimitingChar string, sourceFile string, dataFile string) (summary Summary, err error) {
	summary.DataFile = dataFile
	inputFile, err := os.Open(sourceFile)
	if err != nil {
		return summary, err
	}
	defer func() { _ = inputFile.Close() }()

	outputFile, err := os.Create(dataFile)
	if err != nil {
		return summary, err
	}
	// Capture Close errors so a failed flush on the output index file does
	// not silently produce a truncated/corrupt .dat file.
	defer func() {
		if cerr := outputFile.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close %q: %w", dataFile, cerr)
		}
	}()

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
				if werr := pkg.WriteDataPos(outputFile, int(pkg.DataTableSize), totalFortunes, dataPos); werr != nil {
					return summary, werr
				}
			} else {
				fortuneBase = append(fortuneBase, dataPos)
			}

			fortuneBytes = make([]byte, 0)
		} else {
			fortuneBytes = append(fortuneBytes[:], fortunePortion[:]...)
		}
	}

	if err := scanner.Err(); err != nil {
		return summary, err
	}

	flags := calculateFlags(randomize, order, rot13)
	posContents := pkg.CreateDataTable(totalFortunes, longestFortune, shortestFortune, flags, delimitingChar)
	if err := pkg.SaveDataTable(outputFile, posContents); err != nil {
		return summary, err
	}

	if order {
		sort.Slice(fortuneBase, func(i, j int) bool {
			return pkg.LessThanDataPos(fortuneBase[i], fortuneBase[j])
		})
	} else if randomize {
		Shuffle(fortuneBase)
	}

	if order || randomize {
		if werr := pkg.WriteDataPosSlice(outputFile, int(pkg.DataTableSize), fortuneBase); werr != nil {
			return summary, werr
		}
	}

	summary.TotalFortunes = totalFortunes
	summary.LongestFortune = longestFortune
	summary.ShortestFortune = shortestFortune
	return summary, nil
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
