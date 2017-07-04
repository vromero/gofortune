package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/gofortune/gofortune/lib"
	"github.com/gofortune/gofortune/lib/strfile"
)

type StrFileRequest struct {
	DelimitingChar, SourceFile, DataFile        string
	IgnoreCase, Silent, Order, Randomize, Rot13 bool
}

var strFileCmdRequest = StrFileRequest{}

var strFileName string = "strfile"
var strFileShortDescription string = "Create a random access index file for storing string"
var strFileLongDescription string = `strfile reads a file containing groups of lines separated by a line containing a
single percent '%' sign (or other specified delimiter character) and creates a data file which contains a header
structure and a table of file offsets for each group of lines. This allows random access of the strings.
The output file, if not specified on the command line, is named sourcefile.dat.`

var strfileCmd = &cobra.Command{
	Use:   strFileName,
	Short: strFileShortDescription,
	Long:  strFileLongDescription,
	Run: func(cmd *cobra.Command, args []string) {
		strfilePrepareRequest(args)
	},
}

func init() {
	RootCmd.AddCommand(strfileCmd)
	strfileCmd.Flags().StringVarP(&strFileCmdRequest.DelimitingChar, "delimitingChar", "c", "%", "Change the delimiting character from the percent sign to DelimitingChar")
	strfileCmd.Flags().BoolVarP(&strFileCmdRequest.IgnoreCase, "ignoreCase", "i", false, "Ignore case when ordering the strings")
	strfileCmd.Flags().BoolVarP(&strFileCmdRequest.Silent, "silent", "s", false, "Run silently")
	strfileCmd.Flags().BoolVarP(&strFileCmdRequest.Order, "order", "o", false, "Order the strings in alphabetical Order")
	strfileCmd.Flags().BoolVarP(&strFileCmdRequest.Randomize, "randomize", "n", false, "Randomize  access  to  the strings")
	strfileCmd.Flags().BoolVarP(&strFileCmdRequest.Rot13, "rot13", "x", false, "Rotate  13  positions  in  a simple caesar cypher")
}

func strfilePrepareRequest(args []string) {

	if len(args) < 1 {
		fmt.Println("No input file name")
		os.Exit(1)
	}

	strFileCmdRequest.SourceFile = args[0]

	if len(args[1]) > 0 {
		strFileCmdRequest.DataFile = args[1]
	} else {
		strFileCmdRequest.DataFile = lib.RemoveFileExtension(args[0]) + ".dat"
	}

	strFileRun(strFileCmdRequest)
}

func strFileRun(request StrFileRequest) (err error) {
	return strfile.StrFile(request.IgnoreCase, request.Silent, request.Order, request.Randomize, request.Rot13,
		request.DelimitingChar, request.SourceFile, request.DataFile)
}
