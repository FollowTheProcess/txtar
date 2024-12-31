package txtar_test

import (
	"testing"

	"github.com/FollowTheProcess/test"
	"github.com/FollowTheProcess/txtar"
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
				txtar.WithFile("file1", []byte("some stuff")),
			},
			files:   []string{"file1"},
			wantErr: false,
		},
		{
			name: "multiple unique files",
			options: []txtar.Option{
				txtar.WithFile("file1", []byte("some stuff")),
				txtar.WithFile("file2", []byte("some stuff")),
				txtar.WithFile("file3", []byte("some stuff")),
				txtar.WithFile("file4", []byte("some stuff")),
				txtar.WithFile("file5", []byte("some stuff")),
			},
			files:   []string{"file1", "file2", "file3", "file4", "file5"},
			wantErr: false,
		},
		{
			name: "duplicate file",
			options: []txtar.Option{
				txtar.WithFile("file1", []byte("some stuff")),
				txtar.WithFile("file1", []byte("some different stuff")),
			},
			files:   []string{"file1"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, err := txtar.New(tt.options...)
			test.WantErr(t, err, tt.wantErr)

			if err == nil {
				for _, file := range tt.files {
					test.True(t, archive.Has(file))
				}
			}
		})
	}
}

func TestArchiveAdd(t *testing.T) {
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
				err := archive.Add(file, []byte("some stuff"))
				test.Ok(t, err)
			}
		})
	}
}

func TestArchiveAddDuplicate(t *testing.T) {
	archive, err := txtar.New()
	test.Ok(t, err)

	test.Ok(t, archive.Add("file1", []byte("some stuff")))

	test.Err(
		t,
		archive.Add("file1", []byte("different stuff")),
	) // Did not error on Add duplicate file
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
			options: []txtar.Option{txtar.WithFile("afile", []byte("some stuff"))},
			files: map[string]bool{
				"afile":   true,
				"another": false,
			},
		},
		{
			name: "both",
			options: []txtar.Option{
				txtar.WithFile("afile", []byte("some stuff")),
				txtar.WithFile("another", []byte("moar stuff")),
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
		wantErr bool              // Whether Read should return an error
	}{
		{
			name:    "empty",
			options: nil,
			files:   nil,
			wantErr: false,
		},
		{
			name: "missing",
			options: []txtar.Option{
				txtar.WithFile("exists.txt", []byte("some stuff here")),
			},
			files: map[string]string{
				"missing.txt": "",
			},
			wantErr: true,
		},
		{
			name: "exists",
			options: []txtar.Option{
				txtar.WithFile("exists.txt", []byte("some stuff here")),
			},
			files: map[string]string{
				"exists.txt": "some stuff here",
			},
			wantErr: false,
		},
		{
			name: "exists but empty",
			options: []txtar.Option{
				txtar.WithFile("exists.txt", nil),
			},
			files: map[string]string{
				"exists.txt": "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		archive, err := txtar.New(tt.options...)
		test.Ok(t, err)

		for name, want := range tt.files {
			got, err := archive.Read(name)
			test.WantErr(t, err, tt.wantErr)

			test.Equal(t, string(got), want)
		}
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

		err = archive.Add("present", []byte("present stuff"))
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
				txtar.WithFile("file1.txt", []byte("file1 contents")),
			},
			want: "-- file1.txt --\nfile1 contents\n",
		},
		{
			name: "file and comment",
			options: []txtar.Option{
				txtar.WithComment("A comment"),
				txtar.WithFile("file1.txt", []byte("file1 contents")),
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
				txtar.WithFile("afile.txt", []byte("file1 contents")),
				txtar.WithFile("bfile.txt", []byte("file2 contents")),
				txtar.WithFile("dir/file3.txt", []byte("dir/file3 contents")),
				txtar.WithFile("cfile.txt", []byte("file4 contents")),
				txtar.WithFile("file.txt", []byte("file contents")),
			},
			want: `A slightly longer comment

spanning several
lines

-- afile.txt --
file1 contents
-- bfile.txt --
file2 contents
-- cfile.txt --
file4 contents
-- dir/file3.txt --
dir/file3 contents
-- file.txt --
file contents
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, err := txtar.New(tt.options...)
			test.Ok(t, err)

			test.Equal(t, archive.String(), tt.want)
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		files   map[string]string // The expected files that should exist in the parsed archive
		name    string            // Name of the test case
		input   string            // Input to parse
		errMsg  string            // If it does return an error, what should it say
		comment string            // Expected top level comment of the archive
		wantErr bool              // Whether Parse should return an error
	}{
		{
			name:    "empty",
			input:   "",
			wantErr: true,
			errMsg:  "Parse: cannot parse empty txtar archive",
			comment: "",
			files:   nil,
		},
		{
			name:    "no files",
			input:   "Just a top level comment",
			wantErr: true,
			errMsg:  "Parse: archive contains no files",
			comment: "",
			files:   nil,
		},
		{
			name: "one file no comment",
			input: `-- file1.txt --
file1 contents
`,
			wantErr: false,
			errMsg:  "",
			comment: "",
			files: map[string]string{
				"file1.txt": "file1 contents",
			},
		},
		{
			name: "one file with comment",
			input: `I'm a top level comment
			
-- file1.txt --
file1 contents
`,
			wantErr: false,
			errMsg:  "",
			comment: "I'm a top level comment",
			files: map[string]string{
				"file1.txt": "file1 contents",
			},
		},
		{
			name: "multiple files",
			input: `I'm a top level comment
			
-- file1.txt --
file1 contents
-- file2.txt --
file2 contents
-- file3.txt --
file3 contents
-- file4.txt --
file4 contents
`,
			wantErr: false,
			errMsg:  "",
			comment: "I'm a top level comment",
			files: map[string]string{
				"file1.txt": "file1 contents",
				"file2.txt": "file2 contents",
				"file3.txt": "file3 contents",
				"file4.txt": "file4 contents",
			},
		},
		{
			name: "multiple files whitespace",
			input: `I'm a top level comment


-- file1.txt --

file1 contents


-- file2.txt --
file2 contents


-- file3.txt --
file3 contents

-- file4.txt --
file4 contents
`,
			wantErr: false,
			errMsg:  "",
			comment: "I'm a top level comment",
			files: map[string]string{
				"file1.txt": "file1 contents",
				"file2.txt": "file2 contents",
				"file3.txt": "file3 contents",
				"file4.txt": "file4 contents",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, err := txtar.Parse([]byte(tt.input))
			test.WantErr(t, err, tt.wantErr)

			if err != nil {
				test.Equal(t, err.Error(), tt.errMsg)
			}

			if err == nil {
				test.Equal(t, archive.Comment(), tt.comment) // Comment did not match expected

				for file, contents := range tt.files {
					test.True(t, archive.Has(file)) // Archive was missing file

					got, err := archive.Read(file)
					test.Ok(t, err)

					test.Equal(t, string(got), contents) // File contents differed from expected
				}
			}
		})
	}
}
