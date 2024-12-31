package txtar_test

import (
	"bytes"
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

func TestArchiveHasFile(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		archive, err := txtar.New()
		test.Ok(t, err)

		test.False(t, archive.Has("missing.txt")) // Has() reported true when it shouldn't
	})

	t.Run("using WithFile", func(t *testing.T) {
		archive, err := txtar.New(txtar.WithFile("exists.txt", []byte("contents for exists.txt")))
		test.Ok(t, err)

		test.True(t, archive.Has("exists.txt")) // exists.txt should exist after WithFile
	})

	t.Run("using Add", func(t *testing.T) {
		archive, err := txtar.New()
		test.Ok(t, err)

		err = archive.Add("exists.txt", []byte("contents for exists.txt"))
		test.Ok(t, err)

		test.True(t, archive.Has("exists.txt")) // exists.txt should exist after Add
	})
}

func TestArchiveRead(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		archive, err := txtar.New()
		test.Ok(t, err)

		contents, err := archive.Read("missing")
		test.Err(t, err) // Read did not return error on missing file
		test.EqualFunc(t, contents, nil, bytes.Equal)
	})

	t.Run("empty", func(t *testing.T) {
		archive, err := txtar.New(txtar.WithFile("empty.txt", nil))
		test.Ok(t, err)

		contents, err := archive.Read("empty.txt")
		test.Ok(t, err) // File is empty, not missing, should be no error
		test.EqualFunc(t, contents, nil, bytes.Equal)
	})

	t.Run("full", func(t *testing.T) {
		archive, err := txtar.New(txtar.WithFile("full.txt", []byte("stuff here")))
		test.Ok(t, err)

		contents, err := archive.Read("full.txt")
		test.Ok(t, err)
		test.EqualFunc(t, contents, []byte("stuff here"), bytes.Equal)
	})
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
