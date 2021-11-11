package mlocate

import (
	"encoding/binary"
	"math/rand"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	testDBBytes     = []byte("\x00mlocate\x00\x00\x00\x4E\x00\x01\x00\x00/\x00prune_bind_mounts\x001\x00\x00prunefs\x009P\x00AFS\x00\x00prunenames\x00.git\x00.hg\x00.svn\x00\x00prunepaths\x00/tmp\x00\x00\x00\x00\x00\x00\x57\xE7\x9A\xE0\x07c\x86\x13\x00\x00\x00\x00/\x00\x00bin\x00\x01boot\x00\x02\x00\x00\x00\x00\x61\x8C\x1E\xB2\x07\x5B\xCD\x15\x00\x00\x00\x00/etc\x00\x00foo\x00\x01bar\x00\x02")
	cmpOpts         = cmpopts.IgnoreUnexported(Header{}, DirEntry{}, FileEntry{})
	letters         = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
)

type FileStructure struct {
	Path           string
	Name           string
	Files          []FileStructure
	DirTimeSeconds uint64
	DirTimeNanos   uint32
}

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randInt(max int) int {
	return rand.Intn(max)
}

func randFiles(n int, path ...string) []FileStructure {
	pathPrefix := ""
	if len(path) > 0 {
		pathPrefix = path[0]
	}

	ret := make([]FileStructure, n)

	for i := 0; i < n; i++ {
		ret[i] = FileStructure{
			Name:           randStr(5),
			Path:           pathPrefix,
			DirTimeSeconds: uint64(time.Now().Unix()),
			DirTimeNanos:   uint32(randInt(999999999)),
		}
	}

	return ret
}

func randDirs(n int, maxDepth int, path ...string) []FileStructure {
	pathPrefix := "/"
	if len(path) > 0 {
		pathPrefix = path[0]
	}

	ret := make([]FileStructure, n)

	for i := 0; i < n; i++ {
		fs := &FileStructure{}
		fs.Files = randFiles(2, pathPrefix)

		fs.DirTimeSeconds = uint64(time.Now().Unix())
		fs.DirTimeNanos   = uint32(randInt(999999999))
		fs.Name           = randStr(5)
		fs.Path           = pathPrefix

		if maxDepth > 0 {
			fs.Files = append(fs.Files, randDirs(2, maxDepth - 1, pathPrefix + fs.Name)...)
		}

		ret[i] = *fs
	}

	return ret
}

func (fs FileStructure) toDBFormat(bytes ...byte) []byte {
	seconds  := make([]byte, 8)
	nanos    := make([]byte, 4)
	padding  := []byte{0, 0, 0, 0}
	pathName := []byte(fs.Path + "/" + fs.Name + "\x00")

	binary.BigEndian.PutUint64(seconds, fs.DirTimeSeconds)
	binary.BigEndian.PutUint32(nanos, fs.DirTimeNanos)

	bytes = append(bytes, seconds...)
	bytes = append(bytes, nanos...)
	bytes = append(bytes, padding...)
	bytes = append(bytes, pathName...)

	for _, f := range fs.Files {
		if len(f.Files) > 0 {
			bytes = append(bytes, 1)
		} else {
			bytes = append(bytes, 0)
		}

		bytes = append(bytes, []byte(f.Name)...)
		bytes = append(bytes, 0)
	}

	bytes = append(bytes, 2)

	for _, f := range fs.Files {
		if len(f.Files) > 0 {
			bytes = append(bytes, f.toDBFormat(bytes...)...)
		}
	}

	return bytes
}

func mockDB() DB {
	header := Header{
		MagicNumber:            magicNum,
		ConfigurationBlockSize: 78,
		FileFormatVersion:      0,
		RequireVisibility:      1,
		DatabasePath:           "/",
	}

	configuration := ConfigurationBlock{
		PruneBindMounts: []string{"1"},
		PruneFS:         []string{"9P", "AFS"},
		PruneNames:      []string{".git", ".hg", ".svn"},
		PrunePaths:      []string{"/tmp"},
	}

	dir1 := DirEntry{
		DirTimeSeconds: 1474796256,
		DirTimeNanos:   123962899,
		PathName:       "/",
		Files:          []FileEntry{
			{_type: 0, Name:  "bin"},
			{_type: 1, Name:  "boot"},
		},
	}
	dir2 := DirEntry{
		DirTimeSeconds: 1636572850,
		DirTimeNanos:   123456789,
		PathName:       "/etc",
		Files:          []FileEntry{
			{
				_type: 0,
				Name:  "foo",
			},
			{
				_type: 1,
				Name:  "bar",
			},
		},
	}

	directories := []DirEntry{
		dir1,
		dir2,
	}

	return DB{
		Header:             header,
		ConfigurationBlock: configuration,
		Directories:        directories,
		Index:              map[string]*DirEntry{
			"/":    &dir1,
			"/etc": &dir2,
		},
	}
}

func Test_parseHeader(t *testing.T) {
	got := &DB{}
	want := mockDB().Header
	got.parseHeader(testDBBytes)

	if diff := cmp.Diff(want, got.Header, cmpOpts); diff != "" {
		t.Errorf("parseHeader() mismatch (-want +got):\n%s", diff)
	}
}

func Test_parseConfigurationBlock(t *testing.T) {
	configSize := mockDB().Header.ConfigurationBlockSize

	got := &DB{}
	want := mockDB().ConfigurationBlock
	got.parseConfigurationBlock(testDBBytes, configSize, 17 + uint32(len(mockDB().Header.DatabasePath)))

	if diff := cmp.Diff(want, got.ConfigurationBlock, cmpOpts); diff != "" {
		t.Errorf("parseConfigurationBlock() mismatch (-want +got):\n%s", diff)
	}
}

func Test_parseDirectories(t *testing.T) {
	configSize := mockDB().Header.ConfigurationBlockSize

	got := &DB{}
	want := mockDB().Directories
	got.parseDirectories(testDBBytes, configSize, uint32(len(mockDB().Header.DatabasePath)))

	if diff := cmp.Diff(want, got.Directories, cmpOpts); diff != "" {
		t.Errorf("parseDirectories() mismatch (-want +got):\n%s", diff)
	}
}

func Test_New(t *testing.T) {
	want := mockDB()
	got := New(testDBBytes...)

	if diff := cmp.Diff(want, got, cmpOpts); diff != "" {
		t.Errorf("New() mismatch (-want +got):\n%s", diff)
	}
}

var (
	header  = []byte("\x00mlocate\x00\x00\x00\x4E\x00\x01\x00\x00/\x00prune_bind_mounts\x001\x00\x00prunefs\x009P\x00AFS\x00\x00prunenames\x00.git\x00.hg\x00.svn\x00\x00prunepaths\x00/tmp\x00\x00")
	dirs    = randDirs(1, 2)[0]
	benchDB = dirs.toDBFormat(header...)
)

func Benchmark(b *testing.B) {
	for n := 0; n < b.N; n++ {
		New(benchDB...)
	}
}
