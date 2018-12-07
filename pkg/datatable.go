package pkg

import (
	"bytes"
	"encoding/binary"
	"os"
	"unicode/utf8"
	"unsafe"
)

const (
	DataTableSize = uint64(unsafe.Sizeof(DataTable{}))

	DefaultVersion        = 2
	FlagRandom     uint32 = 1 /* randomized pointers */
	FlagOrdered    uint32 = 2 /* ordered pointers */
	FlagRotated    uint32 = 4 /* rot-13'd text */
)

type DataTableVersion struct {
	Version uint32
}

type DataTable struct {
	Version         uint32
	NumberOfStrings uint32
	LongestLength   uint32
	ShortestLength  uint32
	Flags           uint32
	Delimiter       uint8
	Stuff           [3]uint8
}

func CreateDataTable(numberOfStrings uint32, longestLength uint32, shortestLength uint32, flags uint32, delimiter string) (posContents DataTable) {
	delimiterValue, _ := utf8.DecodeRuneInString(delimiter)
	return DataTable{
		Version:         DefaultVersion,
		NumberOfStrings: numberOfStrings,
		LongestLength:   longestLength,
		ShortestLength:  shortestLength,
		Flags:           flags,
		Delimiter:       uint8(delimiterValue)}
}

func LoadDataTableVersionFromPath(inputFilePath string) (posContents DataTableVersion, err error) {
	inputFile, err := os.Open(inputFilePath)
	defer inputFile.Close()
	return LoadDataTableVersion(inputFile)
}

func LoadDataTableVersion(inputFile *os.File) (posContents DataTableVersion, err error) {
	err = binary.Read(inputFile, binary.BigEndian, &posContents)
	return posContents, err
}

func LoadDataTableFromPath(inputFilePath string) (posContents DataTable, err error) {
	inputFile, err := os.Open(inputFilePath)
	defer inputFile.Close()
	return LoadDataTable(inputFile)
}

func LoadDataTable(inputFile *os.File) (posContents DataTable, err error) {
	err = binary.Read(inputFile, binary.BigEndian, &posContents)
	return posContents, err
}

func SaveDataTable(outputFile *os.File, posContents DataTable) (err error) {
	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.BigEndian, posContents)
	if err != nil {
		return err
	}
	outputFile.WriteAt(buffer.Bytes(), 0)
	return
}
