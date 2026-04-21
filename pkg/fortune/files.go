package fortune

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vromero/gofortune/pkg"
)

// ErrLengthFilterExcluded is returned by loadFilePath when a file is valid
// but excluded because its longest/shortest entry does not satisfy the
// caller's length filter. Callers should treat this as a non-fatal skip.
var ErrLengthFilterExcluded = errors.New("file does not honor the length filter")

// ProbabilityPath pairs a filesystem path with the percentage probability the
// caller wants it to be randomly selected. Path should point only to
// directories containing fortune files or to a fortune file that has a
// sibling ".dat" index.
type ProbabilityPath struct {
	Path       string
	Percentage float32
}

// FileSystemNodeDescriptor is a node in the fortune tree: a directory (with
// Children) or a leaf fortune file (with IndexPath and Table populated).
type FileSystemNodeDescriptor struct {
	Percent                  float32
	UndefinedChildrenPercent float32 // Total percentage non user-defined for this node
	UndefinedNumEntries      uint64
	NumEntries               uint64 // Total number of fortunes in all files
	NumFiles                 int    // Total number of files
	Path                     string
	IndexPath                string
	Table                    pkg.DataTable
	isUtf8                   bool
	Children                 []FileSystemNodeDescriptor
	Parent                   *FileSystemNodeDescriptor
}

// LoadPaths loads the paths described in the paths argument and returns a
// FileSystemNodeDescriptor populated with the tree and each file's index
// table.
//
// LoadPaths can filter fortune files by their shortest/longest dictum, which
// is useful to prevent infinite loops in length-constrained random picks.
func LoadPaths(paths []ProbabilityPath, shorterThan uint32, longerThan uint32) (FileSystemNodeDescriptor, error) {
	rootFsDescriptor := FileSystemNodeDescriptor{
		Percent: 100,
	}

	for i := range paths {
		if err := loadPath(paths[i], &rootFsDescriptor, shorterThan, longerThan); err != nil {
			return rootFsDescriptor, err
		}
	}
	return rootFsDescriptor, nil
}

func loadPath(path ProbabilityPath, parent *FileSystemNodeDescriptor, shorterThan uint32, longerThan uint32) error {
	fsDescriptor := FileSystemNodeDescriptor{
		Path:    path.Path,
		Percent: path.Percentage,
		Parent:  parent,
	}

	stat, err := os.Stat(path.Path)
	if err != nil {
		return fmt.Errorf("stat %q: %w", path.Path, err)
	}
	if stat.IsDir() {
		return loadDirPath(&fsDescriptor, parent, shorterThan, longerThan)
	}
	return loadFilePath(&fsDescriptor, parent, shorterThan, longerThan)
}

func loadDirPath(fsDescriptor *FileSystemNodeDescriptor, parent *FileSystemNodeDescriptor, shorterThan uint32, longerThan uint32) error {
	entries, err := os.ReadDir(fsDescriptor.Path)
	if err != nil {
		return fmt.Errorf("read directory %q: %w", fsDescriptor.Path, err)
	}

	for _, entry := range entries {
		// Sub-directories are ignored for compatibility with the original
		// fortune and because all cookies are typically stored at the top
		// level under /usr/share/games/fortune.
		if entry.IsDir() {
			continue
		}
		childFsDescriptor := FileSystemNodeDescriptor{
			Path:   filepath.Join(fsDescriptor.Path, entry.Name()),
			Parent: fsDescriptor,
		}
		// Files that are not valid fortune files or that fail the length
		// filter are silently skipped for compatibility with fortune(6).
		_ = loadFilePath(&childFsDescriptor, fsDescriptor, shorterThan, longerThan)
	}

	fsDescriptor.Parent = parent
	parent.Children = append(parent.Children, *fsDescriptor)
	return nil
}

func loadFilePath(fsDescriptor *FileSystemNodeDescriptor, parent *FileSystemNodeDescriptor, shorterThan uint32, longerThan uint32) error {
	if !isFortuneFile(fsDescriptor.Path) {
		return fmt.Errorf("%q is not a valid fortune file", fsDescriptor.Path)
	}

	indexPath := fsDescriptor.Path + ".dat"
	if !isFortuneIndexFile(indexPath) {
		return fmt.Errorf("%q is not a valid fortune index file", indexPath)
	}
	fsDescriptor.IndexPath = indexPath

	table, err := pkg.LoadDataTableFromPath(fsDescriptor.IndexPath)
	if err != nil {
		return fmt.Errorf("load data table from %q: %w", fsDescriptor.IndexPath, err)
	}

	if table.LongestLength < longerThan || table.ShortestLength > shorterThan {
		return ErrLengthFilterExcluded
	}

	fsDescriptor.Table = table
	fsDescriptor.Parent = parent

	populateFileAmounts(fsDescriptor, table)
	parent.Children = append(parent.Children, *fsDescriptor)
	return nil
}

func populateFileAmounts(fsDescriptor *FileSystemNodeDescriptor, table pkg.DataTable) {
	current := fsDescriptor
	for current != nil {
		current.NumEntries += uint64(table.NumberOfStrings)
		current.NumFiles++
		current = current.Parent
	}
}

// isFortuneFile reports whether path exists.
func isFortuneFile(path string) bool {
	return pkg.FileExists(path)
}

// isFortuneIndexFile reports whether path is an existing fortune index file
// of the supported version.
func isFortuneIndexFile(path string) bool {
	if !pkg.FileExists(path) {
		return false
	}
	version, err := pkg.LoadDataTableVersionFromPath(path)
	if err != nil || version.Version != pkg.DefaultVersion {
		return false
	}
	return true
}
