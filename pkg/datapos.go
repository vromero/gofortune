package pkg

import (
	"encoding/binary"
	"fmt"
	"os"
	"unsafe"
)

type DataPos struct {
	OriginalOffset uint32
	Text           string
}

func ReadDataPos(inputFile *os.File, tableSize int, position uint32) (DataPos, error) {
	buffer := make([]byte, 4)
	_, err := inputFile.ReadAt(buffer, int64(int64(tableSize)+int64(position)*4))
	if err != nil {
		return DataPos{}, err
	}

	return DataPos{
		OriginalOffset: binary.BigEndian.Uint32(buffer),
	}, nil
}

// WriteDataPos writes a single DataPos entry to outputFile at the offset
// implied by position. Returns any write error so callers can surface a
// failed write instead of silently producing a corrupt index file.
func WriteDataPos(outputFile *os.File, tableSize int, position uint32, datapos DataPos) error {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, datapos.OriginalOffset)
	if _, err := outputFile.WriteAt(buffer, int64(tableSize)+int64(unsafe.Sizeof(position))*int64(position)); err != nil {
		return fmt.Errorf("write data pos at position %d: %w", position, err)
	}
	return nil
}

// WriteDataPosSlice writes every entry in dataposSlice in order. Returns the
// first write error encountered, stopping the iteration.
func WriteDataPosSlice(outputFile *os.File, tableSize int, dataposSlice []DataPos) error {
	for i := range dataposSlice {
		if err := WriteDataPos(outputFile, tableSize, uint32(i), dataposSlice[i]); err != nil {
			return err
		}
	}
	return nil
}

func LessThanDataPos(i DataPos, j DataPos) bool {
	return i.Text[0:1] < j.Text[0:1]
}
