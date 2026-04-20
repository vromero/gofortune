package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vromero/gofortune/pkg"
	"github.com/vromero/gofortune/pkg/strfile"
)

type StrFileRequest struct {
	DelimitingChar, SourceFile, DataFile        string
	IgnoreCase, Silent, Order, Randomize, Rot13 bool
}

var strFileCmdRequest = StrFileRequest{}

var strFileName = "strfile"
var strFileShortDescription = "Create a random access index file for storing string"
var strFileLongDescription = `strfile reads a file containing groups of lines separated by a line containing a
single percent '%' sign (or other specified delimiter character) and creates a data file which contains a header
structure and a table of file offsets for each group of lines. This allows random access of the strings.
The output file, if not specified on the command line, is named sourcefile.dat.`

var strfileCmd = &cobra.Command{
	Use:   strFileName,
	Short: strFileShortDescription,
	Long:  strFileLongDescription,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		strFileCmdRequest.SourceFile = args[0]
		if len(args) > 1 {
			strFileCmdRequest.DataFile = args[1]
		} else {
			strFileCmdRequest.DataFile = pkg.RemoveFileExtension(args[0]) + ".dat"
		}
		return strfile.StrFile(strFileCmdRequest.IgnoreCase, strFileCmdRequest.Silent, strFileCmdRequest.Order, strFileCmdRequest.Randomize, strFileCmdRequest.Rot13,
			strFileCmdRequest.DelimitingChar, strFileCmdRequest.SourceFile, strFileCmdRequest.DataFile)
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

func strFileRun(request StrFileRequest) (err error) {
	return strfile.StrFile(request.IgnoreCase, request.Silent, request.Order, request.Randomize, request.Rot13,
		request.DelimitingChar, request.SourceFile, request.DataFile)
}
