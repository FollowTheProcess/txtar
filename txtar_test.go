package txtar_test

import (
	"bytes"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FollowTheProcess/test"
	"github.com/FollowTheProcess/txtar"
	gotxtar "golang.org/x/tools/txtar"
)

func TestArchiveComment(t *testing.T) {
	tests := []struct {
		name    string // Name of the test case
		comment string // The comment to create the Archive with
		wantErr bool   // Whether New should return an error
	}{
		{
			name:    "empty",
			comment: "",
			wantErr: false,
		},
		{
			name:    "with comment",
			comment: "This is a comment",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, err := txtar.New(txtar.WithComment(tt.comment))
			test.WantErr(t, err, tt.wantErr)

			test.Equal(t, archive.Comment(), tt.comment)
		})
	}
}

func TestWithFiles(t *testing.T) {
	tests := []struct {
		name    string         // Name of the test case
		options []txtar.Option // Options to apply to New
		files   []string       // Filenames that should exist
		wantErr bool           // Whether New should return an error
	}{
		{
			name:    "empty",
			options: nil,
			files:   nil,
			wantErr: false,
		},
		{
			name: "single file",
			options: []txtar.Option{
				txtar.WithFile("file1", "some stuff"),
			},
			files:   []string{"file1"},
			wantErr: false,
		},
		{
			name: "multiple unique files",
			options: []txtar.Option{
				txtar.WithFile("file1", "some stuff"),
				txtar.WithFile("file2", "some stuff"),
				txtar.WithFile("file3", "some stuff"),
				txtar.WithFile("file4", "some stuff"),
				txtar.WithFile("file5", "some stuff"),
			},
			files:   []string{"file1", "file2", "file3", "file4", "file5"},
			wantErr: false,
		},
		{
			name: "duplicate file",
			options: []txtar.Option{
				txtar.WithFile("file1", "some stuff"),
				txtar.WithFile("file1", "some different stuff"),
			},
			files:   []string{"file1"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, err := txtar.New(tt.options...)
			test.WantErr(t, err, tt.wantErr)

			if err == nil {
				for _, file := range tt.files {
					test.True(t, archive.Has(file), test.Context("Missing file %s", file))
				}
			}
		})
	}
}

func TestArchiveWrite(t *testing.T) {
	tests := []struct {
		name  string   // Name of the test case
		files []string // List of files to add (contents don't matter)
	}{
		{
			name:  "empty",
			files: nil,
		},
		{
			name: "single file",
			files: []string{
				"file1",
			},
		},
		{
			name: "multiple files",
			files: []string{
				"file1",
				"file2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, err := txtar.New()
			test.Ok(t, err)

			for _, file := range tt.files {
				err := archive.Write(file, "some stuff")
				test.Ok(t, err)
			}
		})
	}
}

func TestArchiveNilSafe(t *testing.T) {
	var archive *txtar.Archive

	// Everything below must not panic
	err := archive.Write("file", "stuff here")
	test.Err(t, err)
	test.Equal(t, archive.Comment(), "")
	test.False(t, archive.Has("file"))

	contents, ok := archive.Read("file")
	test.False(t, ok)
	test.Equal(t, contents, "")

	test.Equal(t, archive.Size(), 0)
	test.Equal(t, archive.String(), "")
	archive.Delete("file")
	maps.Collect(archive.Files())
}

func TestArchiveWriteDuplicate(t *testing.T) {
	archive, err := txtar.New()
	test.Ok(t, err)

	test.Ok(t, archive.Write("file1", "some stuff"))
	test.Ok(t, archive.Write("file1", "some more stuff"))

	contents, ok := archive.Read("file1")
	test.True(t, ok)
	test.Equal(t, contents, "some more stuff\n")
}

func TestArchiveHas(t *testing.T) {
	tests := []struct {
		name    string          // Name of the test case
		files   map[string]bool // Map of <filename> -> <should exist>
		options []txtar.Option  // The options to pass to New in the test
	}{
		{
			name:    "empty",
			options: nil,
			files:   nil,
		},
		{
			name:    "missing",
			options: []txtar.Option{txtar.WithFile("afile", "some stuff")},
			files: map[string]bool{
				"afile":   true,
				"another": false,
			},
		},
		{
			name: "both",
			options: []txtar.Option{
				txtar.WithFile("afile", "some stuff"),
				txtar.WithFile("another", "moar stuff"),
			},
			files: map[string]bool{
				"afile":   true,
				"another": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, err := txtar.New(tt.options...)
			test.Ok(t, err)

			for name, exists := range tt.files {
				test.Equal(t, archive.Has(name), exists)
			}
		})
	}
}

func TestArchiveRead(t *testing.T) {
	tests := []struct {
		name    string            // Name of the test case
		files   map[string]string // Map of <filename> -> <expected contents> to read
		options []txtar.Option    // The options to apply to New
		wantOk  bool              // Expected Read ok value
	}{
		{
			name:    "empty",
			options: nil,
			files:   nil,
			wantOk:  false,
		},
		{
			name: "missing",
			options: []txtar.Option{
				txtar.WithFile("exists.txt", "some stuff here"),
			},
			files: map[string]string{
				"missing.txt": "",
			},
			wantOk: false,
		},
		{
			name: "exists",
			options: []txtar.Option{
				txtar.WithFile("exists.txt", "some stuff here"),
			},
			files: map[string]string{
				"exists.txt": "some stuff here\n",
			},
			wantOk: true,
		},
		{
			name: "exists but empty",
			options: []txtar.Option{
				txtar.WithFile("exists.txt", ""),
			},
			files: map[string]string{
				"exists.txt": "",
			},
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, err := txtar.New(tt.options...)
			test.Ok(t, err)

			for name, want := range tt.files {
				got, ok := archive.Read(name)
				test.Equal(t, ok, tt.wantOk)

				test.Equal(t, got, want)
			}
		})
	}
}

func TestArchiveDelete(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		archive, err := txtar.New()
		test.Ok(t, err)

		test.False(t, archive.Has("missing"))
		archive.Delete("missing")
		test.False(t, archive.Has("missing"))
	})
	t.Run("present", func(t *testing.T) {
		archive, err := txtar.New()
		test.Ok(t, err)

		test.False(t, archive.Has("present")) // File "present" should not yet be present

		err = archive.Write("present", "present stuff")
		test.Ok(t, err)

		test.True(t, archive.Has("present")) // File "present" should exist
		archive.Delete("present")
		test.False(t, archive.Has("present")) // File "present" should have been deleted
	})
}

func TestArchiveString(t *testing.T) {
	tests := []struct {
		name    string         // Name of the test case
		want    string         // The expected output of calling String
		options []txtar.Option // Options to apply to New
	}{
		{
			name:    "empty",
			options: nil,
			want:    "",
		},
		{
			name: "only comment",
			options: []txtar.Option{
				txtar.WithComment("A comment"),
			},
			want: "A comment\n",
		},
		{
			name: "only single file",
			options: []txtar.Option{
				txtar.WithFile("file1.txt", "file1 contents"),
			},
			want: "-- file1.txt --\nfile1 contents\n",
		},
		{
			name: "file and comment",
			options: []txtar.Option{
				txtar.WithComment("A comment"),
				txtar.WithFile("file1.txt", "file1 contents"),
			},
			want: `A comment

-- file1.txt --
file1 contents
`,
		},
		{
			name: "multiple files",
			options: []txtar.Option{
				txtar.WithComment("A slightly longer comment\n\nspanning several\nlines\n"),
				txtar.WithFile("afile.txt", "file1 contents"),
				txtar.WithFile("bfile.txt", "file2 contents"),
				txtar.WithFile("dir/file3.txt", "dir/file3 contents"),
				txtar.WithFile("cfile.txt", "file4 contents"),
				txtar.WithFile("file.txt", "file contents"),
			},
			want: `A slightly longer comment

spanning several
lines

-- afile.txt --
file1 contents
-- bfile.txt --
file2 contents
-- dir/file3.txt --
dir/file3 contents
-- cfile.txt --
file4 contents
-- file.txt --
file contents
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, err := txtar.New(tt.options...)
			test.Ok(t, err)

			test.Diff(t, archive.String(), tt.want)
		})
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		name        string         // Name of the test case
		thisOptions []txtar.Option // Options to apply to the base archive
		thatOptions []txtar.Option // Options to apply to the other
		want        bool           // What equal should return
	}{
		{
			name:        "empty",
			thisOptions: nil,
			thatOptions: nil,
			want:        true,
		},
		{
			name:        "this comment",
			thisOptions: []txtar.Option{txtar.WithComment("This one")},
			thatOptions: nil,
			want:        false,
		},
		{
			name:        "that comment",
			thisOptions: nil,
			thatOptions: []txtar.Option{txtar.WithComment("That one")},
			want:        false,
		},
		{
			name:        "different comment",
			thisOptions: []txtar.Option{txtar.WithComment("This one")},
			thatOptions: []txtar.Option{txtar.WithComment("That one")},
			want:        false,
		},
		{
			name:        "this empty",
			thisOptions: nil,
			thatOptions: []txtar.Option{
				txtar.WithComment("A comment"),
				txtar.WithFile("file1", "file1 contents"),
			},
			want: false,
		},
		{
			name: "that empty",
			thisOptions: []txtar.Option{
				txtar.WithComment("A comment"),
				txtar.WithFile("file1", "file1 contents"),
				txtar.WithFile("file2", "file2 contents"),
			},
			thatOptions: nil,
			want:        false,
		},
		{
			name: "different len",
			thisOptions: []txtar.Option{
				txtar.WithComment("A comment"),
				txtar.WithFile("file1", "file1 contents"),
				txtar.WithFile("file2", "file2 contents"),
			},
			thatOptions: []txtar.Option{
				txtar.WithComment("A comment"),
				txtar.WithFile("file1", "file1 contents"),
			},
			want: false,
		},
		{
			name: "different filenames",
			thisOptions: []txtar.Option{
				txtar.WithComment("A comment"),
				txtar.WithFile("thisfile1", "file1 contents"),
				txtar.WithFile("file2", "file2 contents"),
			},
			thatOptions: []txtar.Option{
				txtar.WithComment("A comment"),
				txtar.WithFile("thatfile1", "file1 contents"),
				txtar.WithFile("file2", "file2 contents"),
			},
			want: false,
		},
		{
			name: "different contents",
			thisOptions: []txtar.Option{
				txtar.WithFile("file1", "this file1 contents"),
				txtar.WithFile("file2", "this file2 contents"),
			},
			thatOptions: []txtar.Option{
				txtar.WithFile("file1", "that file1 contents"),
				txtar.WithFile("file2", "that file2 contents"),
			},
			want: false,
		},
		{
			name: "equal",
			thisOptions: []txtar.Option{
				txtar.WithComment("A comment"),
				txtar.WithFile("file1", "file1 contents"),
				txtar.WithFile("file2", "file2 contents"),
			},
			thatOptions: []txtar.Option{
				txtar.WithComment("A comment"),
				txtar.WithFile("file1", "file1 contents"),
				txtar.WithFile("file2", "file2 contents"),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this, err := txtar.New(tt.thisOptions...)
			test.Ok(t, err)

			that, err := txtar.New(tt.thatOptions...)
			test.Ok(t, err)

			test.Equal(t, txtar.Equal(this, that), tt.want) // Equal did not return as expected
		})
	}
}

func TestEqualNil(t *testing.T) {
	tests := []struct {
		this *txtar.Archive
		that *txtar.Archive
		name string
		want bool
	}{
		{
			name: "both nil",
			this: nil,
			that: nil,
			want: true,
		},
		{
			name: "this nil",
			this: nil,
			that: &txtar.Archive{},
			want: false,
		},
		{
			name: "that nil",
			this: &txtar.Archive{},
			that: nil,
			want: false,
		},
		{
			name: "both non nil",
			this: &txtar.Archive{},
			that: &txtar.Archive{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.Equal(t, txtar.Equal(tt.this, tt.that), tt.want)
		})
	}
}

func TestParseValid(t *testing.T) {
	tests := []struct {
		files   map[string]string // The expected files that should exist in the parsed archive
		name    string            // Filename of the input file (relative to Testdata/TestParse)
		comment string            // Expected top level comment of the archive
	}{
		{
			name:    "one_file.txtar",
			comment: "",
			files: map[string]string{
				"file1.txt": "file1 contents\n",
			},
		},
		{
			name:    "one_file_with_comment.txtar",
			comment: "I'm a top level comment",
			files: map[string]string{
				"file1.txt": "file1 contents\n",
			},
		},
		{
			name:    "multiple_files.txtar",
			comment: "I'm a top level comment",
			files: map[string]string{
				"file1.txt": "file1 contents\n",
				"file2.txt": "file2 contents\n",
				"file3.txt": "file3 contents\n",
				"file4.txt": "file4 contents\n",
			},
		},
		{
			name:    "multiple_files_whitespace.txtar",
			comment: "I'm a top level comment",
			files: map[string]string{
				"file1.txt": "file1 contents\n",
				"file2.txt": "file2 contents\n",
				"file3.txt": "file3 contents\n",
				"file4.txt": "file4 contents\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", "TestParse", "valid", tt.name)
			file, err := os.Open(path)
			test.Ok(t, err)

			defer file.Close()

			archive, err := txtar.Parse(file)
			test.Ok(t, err) // Parse returned an unexpected error

			test.Equal(t, archive.Comment(), tt.comment) // Comment did not match expected

			for file, contents := range tt.files {
				test.True(t, archive.Has(file), test.Context("Archive was missing %s", file))

				got, ok := archive.Read(file)
				test.True(t, ok)

				test.Equal(t, got, contents, test.Context("Contents differed"))
			}
		})
	}
}

func TestParseFileValid(t *testing.T) {
	tests := []struct {
		files   map[string]string // The expected files that should exist in the parsed archive
		name    string            // Filename of the input file (relative to Testdata/TestParse)
		comment string            // Expected top level comment of the archive
	}{
		{
			name:    "one_file.txtar",
			comment: "",
			files: map[string]string{
				"file1.txt": "file1 contents\n",
			},
		},
		{
			name:    "one_file_with_comment.txtar",
			comment: "I'm a top level comment",
			files: map[string]string{
				"file1.txt": "file1 contents\n",
			},
		},
		{
			name:    "multiple_files.txtar",
			comment: "I'm a top level comment",
			files: map[string]string{
				"file1.txt": "file1 contents\n",
				"file2.txt": "file2 contents\n",
				"file3.txt": "file3 contents\n",
				"file4.txt": "file4 contents\n",
			},
		},
		{
			name:    "multiple_files_whitespace.txtar",
			comment: "I'm a top level comment",
			files: map[string]string{
				"file1.txt": "file1 contents\n",
				"file2.txt": "file2 contents\n",
				"file3.txt": "file3 contents\n",
				"file4.txt": "file4 contents\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", "TestParse", "valid", tt.name)

			archive, err := txtar.ParseFile(path)
			test.Ok(t, err) // Parse returned an unexpected error

			test.Equal(t, archive.Comment(), tt.comment) // Comment did not match expected

			for file, contents := range tt.files {
				test.True(t, archive.Has(file), test.Context("Archive was missing %s", file))

				got, ok := archive.Read(file)
				test.True(t, ok)

				test.Equal(t, got, contents, test.Context("Contents differed"))
			}
		})
	}
}

func TestParseInvalid(t *testing.T) {
	pattern := filepath.Join("testdata", "TestParse", "invalid", "*.txtar")
	files, err := filepath.Glob(pattern)
	test.Ok(t, err) // Could not glob the invalid directory

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			f, err := os.Open(file)
			test.Ok(t, err) // Could not read invalid test case file
			defer f.Close()

			archive, err := txtar.Parse(f)
			test.Err(t, err)            // Parse of invalid file did not return an error
			test.Equal(t, archive, nil) // Archive was not nil
		})
	}
}

func TestParseStringRoundTrip(t *testing.T) {
	pattern := filepath.Join("testdata", "TestParse", "valid", "*.txtar")
	files, err := filepath.Glob(pattern)
	test.Ok(t, err) // Could not glob the valid directory

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			f, err := os.Open(file)
			test.Ok(t, err)
			defer f.Close()

			before, err := txtar.Parse(f)
			test.Ok(t, err) // Could not parse file

			// Stringify it
			stringified := before.String()

			// Reparse it, should be no errors and result in the exact same archive
			after, err := txtar.Parse(strings.NewReader(stringified))
			test.Ok(t, err) // Could not reparse stringified file

			test.Equal(t, before.Comment(), after.Comment()) // Comment mismatch before vs after
			test.Equal(
				t,
				before.Size(),
				after.Size(),
			) // Number of files mismatch before vs after
			test.Equal(t, before.String(), after.String()) // String() mismatch before vs after
		})
	}
}

func TestFiles(t *testing.T) {
	archive, err := txtar.New(
		txtar.WithFile("file1", "some stuff"),
		txtar.WithFile("file2", "file2 stuff"),
		txtar.WithFile("file3", "file3 stuff"),
		txtar.WithFile("file4", "file4 stuff"),
	)
	test.Ok(t, err)

	files := maps.Collect(archive.Files())

	test.Equal(t, len(files), 4, test.Context("Wrong number of files from the iterator"))
	test.Equal(t, files["file1"], "some stuff\n", test.Context("Wrong contents for file1"))
	test.Equal(t, files["file2"], "file2 stuff\n", test.Context("Wrong contents for file2"))
	test.Equal(t, files["file3"], "file3 stuff\n", test.Context("Wrong contents for file3"))
	test.Equal(t, files["file4"], "file4 stuff\n", test.Context("Wrong contents for file4"))
}

func TestCompat(t *testing.T) {
	pattern := filepath.Join("testdata", "TestCompat", "*.txtar")
	files, err := filepath.Glob(pattern)
	test.Ok(t, err) // Could not glob the TestCompat directory

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			// Note: we're not testing we both error in the same conditions
			// because we are intentionally being stricter
			contents, err := os.ReadFile(file)
			test.Ok(t, err)

			// We need to normalise line endings to get equivalent behaviour on all platforms
			contents = bytes.ReplaceAll(contents, []byte("\r\n"), []byte("\n"))

			goArchive := gotxtar.Parse(contents)

			ourArchive, err := txtar.Parse(bytes.NewReader(contents))
			test.Ok(t, err) // our txtar could not parse file

			test.Equal( // Comment mismatch between x/tools/txtar and this package
				t,
				clean(string(goArchive.Comment)),
				strings.TrimSpace(ourArchive.Comment()),
			)

			test.Equal(t, len(goArchive.Files), ourArchive.Size()) // Mismatch in number of files

			for _, file := range goArchive.Files {
				test.True(t, ourArchive.Has(file.Name)) // This package archive missing file
				ourData, ok := ourArchive.Read(file.Name)
				test.True(t, ok) // Could not read data

				test.Equal(t, clean(ourData), clean(string(file.Data))) // File data mismatch
			}
		})
	}
}

func TestDump(t *testing.T) {
	archive, err := txtar.New(
		txtar.WithComment("A top level comment"),
		txtar.WithFile("file1", "file 1 contents"),
		txtar.WithFile("file2", "file 2 contents"),
		txtar.WithFile("file3", "file 3 contents"),
	)

	test.Ok(t, err)

	buf := &bytes.Buffer{}
	err = txtar.Dump(buf, archive)
	test.Ok(t, err)

	test.Equal(t, buf.String(), archive.String())
}

func TestDumpFile(t *testing.T) {
	archive, err := txtar.New(
		txtar.WithComment("A top level comment"),
		txtar.WithFile("file1", "file 1 contents"),
		txtar.WithFile("file2", "file 2 contents"),
		txtar.WithFile("file3", "file 3 contents"),
	)

	test.Ok(t, err)

	tmp, err := os.CreateTemp(t.TempDir(), "*.txtar")
	test.Ok(t, err)
	test.Ok(t, tmp.Close())

	err = txtar.DumpFile(tmp.Name(), archive)
	test.Ok(t, err)

	got, err := os.ReadFile(tmp.Name())
	test.Ok(t, err)

	test.Diff(t, string(got), archive.String())
}

func TestDumpNilSafe(t *testing.T) {
	var archive *txtar.Archive

	buf := &bytes.Buffer{}
	err := txtar.Dump(buf, archive)
	test.Err(t, err)
}

// clean de-windows's everything and trims all leading and trailing whitespace
// returning a byte slice.
func clean(data string) string {
	data = strings.ReplaceAll(data, "\r\n", "\n")

	return strings.TrimSpace(data)
}
