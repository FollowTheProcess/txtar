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
//   - File names and contents are stored with all leading and trailing whitespace trimmed so that formatting the archive is easier and more consistent
//   - Parsing an [Archive] from it's serialised format *can* error in the presence of a malformed document
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
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"
)

var (
	newlineMarker = []byte("\n-- ")
	marker        = []byte("-- ")
	markerEnd     = []byte(" --")
)

// Archive is a collection of files.
//
// Unlike the original package, an Archive's fields are private with access provided by
// an ergonomic API to read, write and delete individual files.
//
// An Archive is not safe for concurrent access across multiple goroutines, the caller
// is responsible for synchronising concurrent access.
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
	name = strings.TrimSpace(name)
	_, exists := a.files[name]
	return exists
}

// Add adds a new named file with contents to the archive.
//
// File names must be unique within an archive so attempting to add a
// duplicate file will result in an error.
//
// The file contents will have leading and trailing whitespace trimmed so that
// formatting can be kept consistent when parsing and serialising an archive.
func (a *Archive) Add(name string, contents []byte) error {
	name = strings.TrimSpace(name)
	if _, exists := a.files[name]; exists {
		return fmt.Errorf("file with name %s already exists in archive", name)
	}

	contents = bytes.TrimSpace(contents)

	a.files[name] = contents
	return nil
}

// Read returns the contents of the given file from the archive.
//
// If the file is not in the archive, an error will be returned.
func (a Archive) Read(name string) ([]byte, error) {
	name = strings.TrimSpace(name)
	contents, exists := a.files[name]
	if !exists {
		return nil, fmt.Errorf("file %s not contained in the archive", name)
	}
	return contents, nil
}

// Delete removes a file from the archive.
//
// If the file does not exist, Delete is a no-op.
func (a *Archive) Delete(name string) {
	name = strings.TrimSpace(name)
	delete(a.files, name)
}

// String implements the [fmt.Stringer] interface for an [Archive], allowing
// it to print itself.
//
// The files will be printed sorted by filename.
func (a Archive) String() string {
	s := &strings.Builder{}

	if a.comment != "" {
		s.WriteString(a.comment)
		s.WriteString("\n")

		// If there are files after the comment we need an extra newline after the comment
		if len(a.files) != 0 {
			s.WriteByte('\n')
		}
	}

	// Sort by filename so the output is deterministic
	names := make([]string, 0, len(a.files))
	for name := range a.files {
		names = append(names, name)
	}
	slices.Sort(names)

	for _, name := range names {
		s.WriteString("-- ")
		s.WriteString(name)
		s.WriteString(" --\n")
		s.Write(a.files[name])
		s.WriteByte('\n')
	}

	return s.String()
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

// Parse constructs an [Archive] from it's serialised representation in text.
//
// Unlike the original txtar package, Parse can (and will) return an error in
// the presence of a malformed document.
func Parse(data []byte) (*Archive, error) {
	// TODO(@FollowTheProcess): This is more or less the original parser just to get things working
	// explore edge cases and see if we can make the errors clear and handle any weird syntax issues

	if len(data) == 0 {
		return nil, errors.New("Parse: cannot parse empty txtar archive")
	}

	if !bytes.Contains(data, marker) {
		return nil, errors.New("Parse: archive contains no files")
	}

	archive := &Archive{
		files: make(map[string][]byte),
	}

	comment, name, data := findFileMarker(data)
	archive.comment = string(bytes.TrimSpace(comment))

	for name != "" {
		fileName := name // Copy of the "before" filename
		var contents []byte
		contents, name, data = findFileMarker(data)
		archive.files[fileName] = bytes.TrimSpace(contents)
	}

	return archive, nil
}

// findFileMarker finds the next file marker in data,
// extracts the file name, and returns the data before the marker,
// the file name, and the data after the marker.
// If there is no next marker, findFileMarker returns before = fixNL(data), name = "", after = nil.
func findFileMarker(data []byte) (before []byte, name string, after []byte) {
	var i int
	for {
		if name, after = isMarker(data[i:]); name != "" {
			return data[:i], name, after
		}
		j := bytes.Index(data[i:], newlineMarker)
		if j < 0 {
			return fixNL(data), "", nil
		}
		i += j + 1 // positioned at start of new possible marker
	}
}

// isMarker checks whether data begins with a file marker line.
// If so, it returns the name from the line and the data after the line.
// Otherwise it returns name == "" with an unspecified after.
func isMarker(data []byte) (name string, after []byte) {
	if !bytes.HasPrefix(data, marker) {
		return "", nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		data, after = data[:i], data[i+1:]
	}
	if !(bytes.HasSuffix(data, markerEnd) && len(data) >= len(marker)+len(markerEnd)) {
		return "", nil
	}
	return strings.TrimSpace(string(data[len(marker) : len(data)-len(markerEnd)])), after
}

// If data is empty or ends in \n, fixNL returns data.
// Otherwise fixNL returns a new slice consisting of data with a final \n added.
func fixNL(data []byte) []byte {
	if len(data) == 0 || data[len(data)-1] == '\n' {
		return data
	}
	d := make([]byte, len(data)+1)
	copy(d, data)
	d[len(data)] = '\n'
	return d
}
