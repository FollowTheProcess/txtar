package txtar

import "strings"

// Option is a functional option for building/configuring an [Archive].
type Option func(*Archive) error

// WithComment is an [Option] that sets the top level comment for an [Archive].
//
// Leading and trailing whitespace is stripped from the comment before adding so that
// the formatting is consistent when printing an archive.
//
// Successive calls overwrite any previous comment.
func WithComment(comment string) Option {
	return func(a *Archive) error {
		a.comment = strings.TrimSpace(comment)

		return nil
	}
}

// WithFile is an [Option] that adds a file to an [Archive].
//
// It is useful for ergonomically building a new archive from Go code
// e.g. in tests.
//
// File names must be unique in an archive, adding a file whose name is already present
// in the archive will cause [New] to return an error.
func WithFile(name string, contents []byte) Option {
	return func(a *Archive) error {
		return a.Add(name, contents)
	}
}
