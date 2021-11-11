package mlocate

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	testDBBytes     = []byte("\x00mlocate\x00\x00\x00\x4E\x00\x01\x00\x00/\x00prune_bind_mounts\x001\x00\x00prunefs\x009P\x00AFS\x00\x00prunenames\x00.git\x00.hg\x00.svn\x00\x00prunepaths\x00/tmp\x00\x00\x00\x00\x00\x00\x57\xE7\x9A\xE0\x07c\x86\x13\x00\x00\x00\x00/\x00\x00bin\x00\x01boot\x00\x02\x00\x00\x00\x00\x61\x8C\x1E\xB2\x07\x5B\xCD\x15\x00\x00\x00\x00/etc\x00\x00foo\x00\x01bar\x00\x02")
	cmpOpts         = cmpopts.IgnoreUnexported(Header{}, DirEntry{}, FileEntry{})
)

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
