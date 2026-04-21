package pkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
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

func LoadDataTableVersionFromPath(inputFilePath string) (DataTableVersion, error) {
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		return DataTableVersion{}, err
	}
	defer func() { _ = inputFile.Close() }()
	return LoadDataTableVersion(inputFile)
}

func LoadDataTableVersion(inputFile *os.File) (posContents DataTableVersion, err error) {
	err = binary.Read(inputFile, binary.BigEndian, &posContents)
	return posContents, err
}

func LoadDataTableFromPath(inputFilePath string) (DataTable, error) {
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		return DataTable{}, err
	}
	defer func() { _ = inputFile.Close() }()
	return LoadDataTable(inputFile)
}

func LoadDataTable(inputFile *os.File) (posContents DataTable, err error) {
	err = binary.Read(inputFile, binary.BigEndian, &posContents)
	return posContents, err
}

// SaveDataTable writes the header posContents at offset 0 of outputFile.
// Returns any encoding or write error so callers can surface a corrupt or
// truncated index instead of silently continuing.
func SaveDataTable(outputFile *os.File, posContents DataTable) error {
	buffer := new(bytes.Buffer)
	if err := binary.Write(buffer, binary.BigEndian, posContents); err != nil {
		return fmt.Errorf("encode data table: %w", err)
	}
	if _, err := outputFile.WriteAt(buffer.Bytes(), 0); err != nil {
		return fmt.Errorf("write data table: %w", err)
	}
	return nil
}
