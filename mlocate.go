package mlocate

import (
	"encoding/binary"
	"errors"
	"reflect"
	"strings"
)

const (
	NUL = "\x00"
	magicNum = NUL + "mlocate"
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

func parseHeader(dbBytes []byte) Header {
	return Header{
		MagicNumber:            string(dbBytes[:8]),
		ConfigurationBlockSize: binary.BigEndian.Uint32(dbBytes[8:13]),
		FileFormatVersion:      uint8(dbBytes[12]),
		RequireVisibility:      uint8(dbBytes[13]),
		DatabasePath:           "/",
	}
}

func parseConfigurationBlock(dbBytes []byte, blockSize uint32, startIndex uint32) ConfigurationBlock {
	configMap := make(map[string][]string)

	config := strings.Split(string(dbBytes[startIndex:startIndex + blockSize]), "\x00\x00")

	for i := 0; i < len(config); i++ {
		split := strings.Split(config[i], "\x00")
		varName := split[0]
		varVals := split[1:]

		configMap[varName] = varVals
	}

	ret := &ConfigurationBlock{}

	ct := reflect.TypeOf(*ret)

	for i := 0; i < ct.NumField(); i++ {
		field := ct.Field(i)
		varName := field.Tag.Get("param")

		switch varName {
		case "prune_bind_mounts":
			ret.PruneBindMounts = configMap[varName]
		case "prunefs":
			ret.PruneFS = configMap[varName]
		case "prunenames":
			ret.PruneNames = configMap[varName]
		case "prunepaths":
			ret.PrunePaths = configMap[varName]
		}
	}

	return *ret
}

func parseDirectories(dbBytes []byte, configBlockSize uint32, pathSize uint32) []DirEntry {
	ret := make([]DirEntry, 0)

	directories := strings.Split(string(dbBytes[16 + configBlockSize + pathSize + 3:]), "\x02")

	for _, d := range directories {
		if d == "" {
			break
		}

		ret = append(ret, parseDirectory([]byte(d)))
	}

	return ret
}

func parseDirectory(dir []byte) DirEntry {
	ret := &DirEntry{}

	pathBytes := make([]byte, 0)
	for i := 16; true; i++ {
		if dir[i] > 0 {
			pathBytes = append(pathBytes, dir[i])
		} else {
			break
		}
	}

	ret.DirTimeSeconds = binary.BigEndian.Uint64(dir[0:8])
	ret.DirTimeNanos = binary.BigEndian.Uint32(dir[8:12])
	ret.PathName = string(pathBytes)
	ret.Files = parseFiles(dir[17 + len(pathBytes):])

	return *ret
}

func parseFiles(fBytes []byte) []FileEntry {
	ret := make([]FileEntry, 0)
	fe := make([]byte, 0)
	for _, b := range fBytes {
		if len(fe) == 0 {
			fe = append(fe, b)
		} else if b == 0 {
			fileEntry := FileEntry{
				_type: uint(fe[0]),
				Name:  string(fe[1:]),
			}
			ret = append(ret, fileEntry)
			fe = make([]byte, 0)
		} else {
			fe = append(fe, b)
		}
	}

	return ret
}

	ret := DB{}

	return ret
}
func New(db ...byte) DB {
