// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package txtar re-implements and extends the original [golang.org/x/tools/txtar] package, making a number
// of modifications and (hopefully) improvements to the package.
//
// No modifications are made to the txtar syntax, all txtar archives produced with this package are
// compatible with the original.
//
// Improvements include:
//
//   - A CLI (cmd/txtar) that can "unzip" and "zip" a txtar archive to the real filesystem (and vice versa)
//   - Files stored in the archive may may be looked up by name and operated on individually
//   - Methods and functions are provided to help easily facilitate individual file editing
//   - A number of useful interfaces are implemented to make an [Archive] more useful/flexible in the Go ecosystem
//   - An ergonomic API for constructing an [Archive], rather than simply exposing struct fields
//
// # Original Package Documentation
//
// Package txtar implements a trivial text-based file archive format.
//
// The goals for the format are:
//
//   - be trivial enough to create and edit by hand.
//   - be able to store trees of text files describing go command test cases.
//   - diff nicely in git history and code reviews.
//
// Non-goals include being a completely general archive format,
// storing binary data, storing file modes, storing special files like
// symbolic links, and so on.
//
// # Txtar format
//
// A txtar archive is zero or more comment lines and then a sequence of file entries.
// Each file entry begins with a file marker line of the form "-- FILENAME --"
// and is followed by zero or more file content lines making up the file data.
// The comment or file content ends at the next file marker line.
// The file marker line must begin with the three-byte sequence "-- "
// and end with the three-byte sequence " --", but the enclosed
// file name can be surrounding by additional white space,
// all of which is stripped.
//
// If the txtar file is missing a trailing newline on the final line,
// parsers should consider a final newline to be present anyway.
//
// There are no possible syntax errors in a txtar archive.
//
// [golang.org/x/tools/txtar]: https://pkg.go.dev/golang.org/x/tools/txtar
package txtar

import (
	"errors"
	"fmt"
)

// Archive is a collection of files.
type Archive struct {
	files   map[string][]byte // The files contained in the archive, map of name to contents
	comment string            // The top level archive comment section
}

// Comment returns the top level archive comment.
func (a Archive) Comment() string {
	return a.comment
}

// Has returns whether the archive contains a file with the given name.
func (a Archive) Has(name string) bool {
	_, exists := a.files[name]
	return exists
}

// Add adds a new named file with contents to the archive.
//
// File names must be unique within an archive so attempting to add a
// duplicate file will result in an error.
func (a *Archive) Add(name string, contents []byte) error {
	if _, exists := a.files[name]; exists {
		return fmt.Errorf("file with name %s already exists in archive", name)
	}

	a.files[name] = contents
	return nil
}

// Read returns the contents of the given file from the archive.
//
// If the file is not in the archive, an error will be returned.
func (a Archive) Read(name string) ([]byte, error) {
	contents, exists := a.files[name]
	if !exists {
		return nil, fmt.Errorf("file %s not contained in the archive", name)
	}
	return contents, nil
}

// New returns a new [Archive], applying any number of initialisation options.
func New(options ...Option) (*Archive, error) {
	archive := &Archive{
		files: make(map[string][]byte),
	}

	// Bubble up all the errors at once rather than forcing callers
	// to play whack-a-mole
	var errs error
	for _, option := range options {
		errs = errors.Join(errs, option(archive))
	}

	if errs != nil {
		return nil, errs
	}

	return archive, nil
}
