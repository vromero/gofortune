package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vromero/gofortune/cmd"
)

func main() {
	processAliases()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// To maintain compatibility with the classic tools: fortune, strfile. GoFortune supports to have its
// executable renamed or symlinked. If the appropriate executable names are found, the command line will
// be altered to honor the original syntax.
func processAliases() {
	switch getExecutableName() {
	case "strfile":
		os.Args = append([]string{"gofortune", "strfile"}, os.Args[1:]...)
	}
}

// Finds the name of the executable (or symlink to executable) used to call this program
func getExecutableName() string {
	return filepath.Base(os.Args[0])
}
