package mlocate

import (
	"encoding/binary"
	"errors"
)

const (
	null = "\x00"
	magicNum = null + "mlocate"
)

type Header struct {
	MagicNumber            string // 8 byte magic number
	ConfigurationBlockSize uint32 // The size of the configuration block
	FileFormatVersion      uint8  // Number indicating file version
	RequireVisibility      uint8  // Whether to check user permissions before reporting results
	DatabasePath           string // Path name of the root of the database
}

type ConfigurationBlock struct {
	PruneBindMounts []string `param:"prune_bind_mounts"` // A  single  entry,  the  value of PRUNE_BIND_MOUNTS; one of the strings 0 or 1
	PruneFS         []string `param:"prunefs"`           // The value of PRUNEFS, each entry is converted to uppercase
	PruneNames      []string `param:"prunenames"`
	PrunePaths      []string `param:"prunepaths"`        // The value of PRUNEPATHS
}

type DB struct {
	Header             Header             // The file header of the database
	ConfigurationBlock ConfigurationBlock // Ensure databases are not reused if some configuration changes  could  affect their  contents
	Directories        []DirEntry         // Entries describing directories and their contents
}

type DirEntry struct {
	DirTimeSeconds uint64      // Maximum of st_ctime and st_mtime in seconds
	DirTimeNanos   uint32      // Nanosecond part of maximum of st_ctime and st_mtime
	PathName       string      // Path name of the directory
	Files          []FileEntry // Sequence of file entries constituting the directory's contents
}

type FileEntry struct {
	_type uint   // 0 (non-directory), 1 (subdirectory), or 2 (end of current directory)
	Name  string // If file entry is a non-directory file or subdirectory, Name is the file's name without its path
}

// Type returns a human friendly string representation of a FileEntry type.
func (fe *FileEntry) Type() (string, error) {
	switch fe._type {
	case 0:
		return "file", nil
	case 1:
		return "subdirectory", nil
	case 2:
		return "end", nil
	default:
		return "", errors.New("invalid file type specification")
	}
}

func New(db ...[]byte) DB {
	ret := DB{}

	return ret
}
