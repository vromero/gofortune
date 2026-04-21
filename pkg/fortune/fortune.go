// Package fortune provides fortune cookie selection. This package will not output
// any data to the terminal.
package fortune

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/patrickdappollonio/localized"
	"github.com/vromero/gofortune/pkg"
)

// maxLengthFilterAttempts bounds the number of random picks
// GetLengthFilteredRandomFortune is willing to make before giving up, to avoid
// infinite loops when no fortune satisfies the length constraint.
const maxLengthFilterAttempts = 1000

// Cookie is a single fortune cookie together with the file it came from.
type Cookie struct {
	Data     string
	FileName string
}

// Request describes a fortune-selection request as produced by PrepareRequest
// and enriched with command-line flags by the caller.
type Request struct {
	AllMaxims, ShowCookieFile, PrintListOfFiles bool
	LongDictumsOnly, ShortOnly, IgnoreCase      bool
	Wait, ConsiderAllEqual, Offensive           bool
	Match                                       string
	LongestShort                                int
	Paths                                       []ProbabilityPath
	OffensivePaths                              []ProbabilityPath
}

// PrepareRequest builds a Request from positional arguments. With no args it
// falls back to the supplied default paths (optionally locale-suffixed). With
// args it parses alternating "N%" and path tokens.
//
// Returns an error if a "N%" token cannot be parsed as an integer.
func PrepareRequest(args []string, defaultFortunePath, defaultOffensiveFortunePath string) (Request, error) {
	request := Request{}
	if len(args) == 0 {
		lang := localized.New()
		_ = lang.Detect()
		lp := filepath.Join(defaultFortunePath, lang.Lang)
		op := filepath.Join(defaultOffensiveFortunePath, lang.Lang)

		request.Paths = []ProbabilityPath{{Path: selectExisting(defaultFortunePath, lp)}}
		request.OffensivePaths = []ProbabilityPath{{Path: selectExisting(defaultOffensiveFortunePath, op)}}
		return request, nil
	}

	currentPath := ProbabilityPath{}
	for _, arg := range args {
		if strings.HasSuffix(arg, "%") {
			raw := strings.TrimSuffix(arg, "%")
			value, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return Request{}, fmt.Errorf("invalid percentage %q: %w", arg, err)
			}
			currentPath.Percentage = float32(value)
			continue
		}
		currentPath.Path = arg
		request.Paths = append(request.Paths, currentPath)
		currentPath = ProbabilityPath{}
	}
	return request, nil
}

func selectExisting(path1 string, path2 string) string {
	if path2 == "" || !pkg.FileExists(path2) {
		return path1
	}
	return path2
}

// GetRandomFortune picks one fortune from a random leaf of the descriptor tree.
func GetRandomFortune(rootNode FileSystemNodeDescriptor) (Cookie, error) {
	randomNode, err := GetRandomLeafNode(rootNode)
	if err != nil {
		return Cookie{}, err
	}
	if randomNode.NumEntries == 0 {
		return Cookie{}, fmt.Errorf("fortune file %q is empty", randomNode.Path)
	}
	randomEntry := rand.IntN(int(randomNode.NumEntries))

	indexFile, err := os.Open(randomNode.IndexPath)
	if err != nil {
		return Cookie{}, fmt.Errorf("open index file %q: %w", randomNode.IndexPath, err)
	}
	defer indexFile.Close()

	dataPos, err := pkg.ReadDataPos(indexFile, int(pkg.DataTableSize), uint32(randomEntry))
	if err != nil {
		return Cookie{}, fmt.Errorf("read index file %q: %w", randomNode.IndexPath, err)
	}

	fortuneFile, err := os.Open(randomNode.Path)
	if err != nil {
		return Cookie{}, fmt.Errorf("open fortune file %q: %w", randomNode.Path, err)
	}
	defer fortuneFile.Close()

	data, err := pkg.ReadData(fortuneFile, int64(dataPos.OriginalOffset))
	if err != nil {
		return Cookie{}, fmt.Errorf("read fortune file %q: %w", randomNode.Path, err)
	}

	return Cookie{FileName: filepath.Base(randomNode.Path), Data: data}, nil
}

// GetLengthFilteredRandomFortune picks a random fortune whose length is in the
// open interval (longerThan, shorterThan). It gives up after
// maxLengthFilterAttempts tries, returning an error, to avoid infinite loops
// when no fortune satisfies the constraint.
func GetLengthFilteredRandomFortune(rootNode FileSystemNodeDescriptor, shorterThan uint32, longerThan uint32) (Cookie, error) {
	for i := 0; i < maxLengthFilterAttempts; i++ {
		cookie, err := GetRandomFortune(rootNode)
		if err != nil {
			return Cookie{}, err
		}
		length := uint32(len(cookie.Data))
		if length > longerThan && length < shorterThan {
			return cookie, nil
		}
	}
	return Cookie{}, fmt.Errorf("no fortune found matching length constraints after %d attempts", maxLengthFilterAttempts)
}

// GetFortunesMatching streams all fortunes matching expression. Returns a data
// channel and an error channel; both are closed when iteration completes.
//
// The implementation walks the descriptor tree iteratively in a single
// goroutine to avoid the goroutine-per-directory explosion (and potential
// goroutine leaks when the consumer stops reading early) of the previous
// design.
func GetFortunesMatching(fsDescriptor FileSystemNodeDescriptor, expression string, ignoreCase bool) (<-chan Cookie, <-chan error) {
	if ignoreCase {
		expression = "(?i)" + expression
	}
	matchingExpression := regexp.MustCompile(expression)
	return getFortunesMatching(fsDescriptor, matchingExpression)
}

func getFortunesMatching(fsDescriptor FileSystemNodeDescriptor, expression *regexp.Regexp) (<-chan Cookie, <-chan error) {
	output := make(chan Cookie)
	errorOutput := make(chan error)

	go func() {
		defer close(output)
		defer close(errorOutput)
		walkMatching(fsDescriptor, expression, output, errorOutput)
	}()

	return output, errorOutput
}

// walkMatching iterates the descriptor tree depth-first, emitting matching
// fortunes to output and failures to errorOutput. Directories with children
// are recursed into; leaves are scanned once.
func walkMatching(node FileSystemNodeDescriptor, expression *regexp.Regexp, output chan<- Cookie, errorOutput chan<- error) {
	if len(node.Children) > 0 {
		for i := range node.Children {
			walkMatching(node.Children[i], expression, output, errorOutput)
		}
		return
	}
	scanLeafForMatches(node, expression, output, errorOutput)
}

func scanLeafForMatches(node FileSystemNodeDescriptor, expression *regexp.Regexp, output chan<- Cookie, errorOutput chan<- error) {
	indexFile, err := os.Open(node.IndexPath)
	if err != nil {
		errorOutput <- fmt.Errorf("open index file %q: %w", node.IndexPath, err)
		return
	}
	defer indexFile.Close()

	fortuneFile, err := os.Open(node.Path)
	if err != nil {
		errorOutput <- fmt.Errorf("open fortune file %q: %w", node.Path, err)
		return
	}
	defer fortuneFile.Close()

	fileName := filepath.Base(node.Path)
	for i := int64(0); i < int64(node.NumEntries); i++ {
		dataPos, err := pkg.ReadDataPos(indexFile, int(pkg.DataTableSize), uint32(i))
		if err != nil {
			errorOutput <- fmt.Errorf("read index file %q entry %d: %w", node.IndexPath, i, err)
			continue
		}

		data, err := pkg.ReadData(fortuneFile, int64(dataPos.OriginalOffset))
		if err != nil {
			errorOutput <- fmt.Errorf("read fortune file %q entry %d: %w", node.Path, i, err)
			continue
		}
		if expression.MatchString(data) {
			output <- Cookie{FileName: fileName, Data: data}
		}
	}
}

// Sentinel errors retained for callers that may want to programmatically
// detect them (not used internally; kept for API surface).
var (
	ErrEmptyFortuneFile = errors.New("fortune file is empty")
)
