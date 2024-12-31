package txtar_test

import (
	"bytes"
	"testing"

	"github.com/FollowTheProcess/test"
	"github.com/FollowTheProcess/txtar"
)

func TestArchiveComment(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		archive, err := txtar.New()
		test.Ok(t, err)

		test.Equal(t, archive.Comment(), "") // Archive comment should be empty
	})

	t.Run("with comment", func(t *testing.T) {
		archive, err := txtar.New(txtar.WithComment("This is a comment"))
		test.Ok(t, err)

		test.Equal(t, archive.Comment(), "This is a comment")
	})
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
