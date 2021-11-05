package mlocate

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	testDBBytes     = []byte("\x00mlocate\x00\x00\x00\x4E\x00\x01\x00\x00/\x00prune_bind_mounts\x001\x00\x00prunefs\x009P\x00AFS\x00\x00prunenames\x00.git\x00.hg\x00.svn\x00\x00prunepaths\x00/tmp\x00\x00\x00\x00\x00\x00\x57\xE7\x9A\xE0\x07c\x86\x13\x00\x00\x00\x00/\x00\x00bin\x00\x01boot\x02\x00")
	cmpOpts         = cmpopts.IgnoreUnexported(Header{}, DirEntry{})
)

func mockDB() *DB {
	header := Header{
		MagicNumber:            magicNum,
		ConfigurationBlockSize: 78,
		FileFormatVersion:      0,
		RequireVisibility:      1,
		padding:                []byte{0, 0},
		DatabasePath:           "/",
	}

	configuration := ConfigurationBlock{
		PruneBindMounts: []string{"1"},
		PruneFS:         []string{"9P", "AFS"},
		PruneNames:      []string{".git", ".hg", ".svn"},
		PrunePaths:      []string{"/tmp"},
	}

	directories := []DirEntry{
		{
			DirTimeSeconds: 1474796256,
			DirTimeNanos:   123962899,
			padding:        []byte{0, 0, 0, 0},
			PathName:       "/",
			Files:          []FileEntry{
				{_type: 0, Name:  "bin"},
				{_type: 1, Name:  "etc"},
			},
		},
	}

	return &DB{
		Header:             header,
		ConfigurationBlock: configuration,
		Directories:        directories,
	}
}

func Test_parseHeader(t *testing.T) {
	want := mockDB().Header
	got := parseHeader(testDBBytes)

	if diff := cmp.Diff(want, got, cmpOpts); diff != "" {
		t.Errorf("parseHeader() mismatch (-want +got):\n%s", diff)
	}
}

func Test_New(t *testing.T) {
	want := mockDB()
	got := New(testDBBytes)

	if diff := cmp.Diff(want, got, cmpOpts); diff != "" {
		t.Errorf("New() mismatch (-want +got):\n%s", diff)
	}
}
