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
//   - Files stored in the archive may may be looked up by name and operated on individually
//   - Methods and functions are provided to help easily facilitate individual file editing
//   - An ergonomic API for constructing an [Archive], rather than simply exposing struct fields
//   - File names and contents are stored with all leading and trailing whitespace trimmed so that formatting the archive is easier and more consistent
//   - Parsing an [Archive] from it's serialised format *can* error in the presence of a malformed document
//   - [Parse] accepts an [io.Reader] rather than a []byte
//   - [Dump] is provided to serialise an [Archive] to an [io.Writer]
//   - File contents are represented as strings, not []byte for a more convenient format
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
	"io"
	"iter"
	"os"
	"strings"

	"github.com/FollowTheProcess/collections/orderedmap"
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
	files   *orderedmap.Map[string, string] // The files contained in the archive, map of name to contents
	comment string                          // The top level archive comment section
}

// Comment returns the top level archive comment.
func (a *Archive) Comment() string {
	if a == nil {
		return ""
	}

	return a.comment
}

// Has returns whether the archive contains a file with the given name.
func (a *Archive) Has(name string) bool {
	if a == nil {
		return false
	}

	a.init()

	name = strings.TrimSpace(name)

	return a.files.Contains(name)
}

// Write writes a named file with contents to the archive.
//
// Calling Write with the name of a file that already exists in the archive will
// overwrite the contents of that file.
//
// The file contents will have leading and trailing whitespace trimmed so that
// formatting can be kept consistent when parsing and serialising an archive.
func (a *Archive) Write(name, contents string) error {
	if a == nil {
		return errors.New("Write called on a nil Archive")
	}

	a.init()

	name = strings.TrimSpace(name)
	contents = strings.TrimSpace(contents)

	a.files.Insert(name, fixNL(contents))

	return nil
}

// Read returns the contents of the given file from the archive.
//
// If the file is not in the archive, an error will be returned.
func (a *Archive) Read(name string) (value string, ok bool) {
	if a == nil {
		return "", false
	}

	a.init()

	name = strings.TrimSpace(name)
	contents, exists := a.files.Get(name)

	if !exists {
		return "", false
	}

	return contents, true
}

// Delete removes a file from the archive.
//
// If the file does not exist, Delete is a no-op.
func (a *Archive) Delete(name string) {
	if a == nil {
		return
	}

	a.init()

	name = strings.TrimSpace(name)
	a.files.Remove(name)
}

// Size returns the number of files stored in the archive.
func (a *Archive) Size() int {
	if a == nil {
		return 0
	}

	a.init()

	return a.files.Size()
}

// String implements the [fmt.Stringer] interface for an [Archive], allowing
// it to print itself.
//
// The files will be printed sorted by filename.
func (a *Archive) String() string {
	if a == nil {
		return ""
	}

	a.init()

	s := &strings.Builder{}

	if a.comment != "" {
		s.WriteString(a.comment)
		s.WriteString("\n")

		// If there are files after the comment we need an extra newline after the comment
		if a.files.Size() != 0 {
			s.WriteByte('\n')
		}
	}

	for name, file := range a.files.All() {
		s.WriteString("-- ")
		s.WriteString(name)
		s.WriteString(" --\n")
		s.WriteString(file)
	}

	return s.String()
}

// Files returns a iterator over the archive's filenames and contents.
//
// The order of iteration is non-deterministic, if order is required the caller
// must collect and sort the results.
func (a *Archive) Files() iter.Seq2[string, string] {
	if a == nil {
		return func(yield func(string, string) bool) {}
	}
	a.init()

	return a.files.All()
}

// init ensures the underlying map is correctly initialised.
func (a *Archive) init() {
	if a.files == nil {
		a.files = orderedmap.New[string, string]()
	}
}

// New returns a new [Archive], applying any number of initialisation options.
func New(options ...Option) (*Archive, error) {
	archive := &Archive{
		files: orderedmap.New[string, string](),
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
// the presence of a malformed document. We also take an [io.Reader] rather than
// a byte slice for greater flexibility.
//
// For a shortcut to parse from a file see [ParseFile].
func Parse(r io.Reader) (*Archive, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.New("Parse: cannot parse empty txtar archive")
	}

	if !bytes.Contains(data, marker) {
		return nil, errors.New("Parse: archive contains no files")
	}

	// Stupid windows
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))

	archive := &Archive{
		files: orderedmap.New[string, string](),
	}

	comment, name, data := findFileMarker(data)
	if data == nil {
		return nil, errors.New("Parse: unterminated file marker")
	}

	archive.comment = string(bytes.TrimSpace(comment))

	for name != "" {
		fileName := name // Copy of the "before" filename

		var contents []byte
		contents, name, data = findFileMarker(data)
		archive.files.Insert(fileName, fixNL(strings.TrimSpace(string(contents))))
	}

	return archive, nil
}

// ParseFile is a convenience wrapper around [Parse] when reading an
// archive from a File.
func ParseFile(name string) (*Archive, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Parse(file)
}

// Dump writes the [Archive] to w in it's serialised representation.
func Dump(w io.Writer, archive *Archive) error {
	if archive == nil {
		return errors.New("Dump: archive was nil")
	}

	_, err := w.Write([]byte(archive.String()))

	return err
}

// DumpFile is a convenience wrapper around [Dump] when serialising an
// archive to a file.
//
// If the file does not exist, it is created.
func DumpFile(name string, archive *Archive) error {
	const filePerms = 0o644
	return os.WriteFile(name, []byte(archive.String()), filePerms)
}

// Equal returns whether two archives should be considered equal.
//
// An archive is considered equal to another if they have the same comment
// and the same files, or they are both nil pointers.
//
// Otherwise they are considered not equal.
func Equal(a, b *Archive) bool {
	// Mirroring the behaviour of maps.Equal, two nil maps report true
	if (a == nil) && (b == nil) {
		return true
	}

	// Can't go any further without bailing if one is nil because
	// it will panic as soon as we implicitly dereference the pointer
	// by calling .comment, .files etc.
	if a == nil || b == nil {
		return false
	}

	// Likewise the underlying maps, mirror maps.Equal
	if (a.files == nil) && (b.files == nil) {
		return true
	}

	// Must also bail if either are nil
	if a.files == nil || b.files == nil {
		return false
	}

	if a.comment != b.comment {
		return false
	}

	if a.files.Size() != b.files.Size() {
		return false
	}

	for file, aContents := range a.files.All() {
		bContents, exists := b.files.Get(file)
		if !exists {
			return false
		}

		if aContents != bContents {
			return false
		}
	}

	return true
}

// Below is verbatim the parser from the original package, the only exception being we don't need fixNL
// because we are stricter about how we store comment and file contents

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
			return data, "", nil
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
func fixNL(data string) string {
	if len(data) == 0 || data[len(data)-1] == '\n' {
		return data
	}
	d := make([]byte, len(data)+1)
	copy(d, data)
	d[len(data)] = '\n'

	return string(d)
}
