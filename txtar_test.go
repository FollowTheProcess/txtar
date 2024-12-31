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
