# txtar

[![License](https://img.shields.io/github/license/FollowTheProcess/txtar)](https://github.com/FollowTheProcess/txtar)
[![Go Report Card](https://goreportcard.com/badge/github.com/FollowTheProcess/txtar)](https://goreportcard.com/report/github.com/FollowTheProcess/txtar)
[![GitHub](https://img.shields.io/github/v/release/FollowTheProcess/txtar?logo=github&sort=semver)](https://github.com/FollowTheProcess/txtar)
[![CI](https://github.com/FollowTheProcess/txtar/workflows/CI/badge.svg)](https://github.com/FollowTheProcess/txtar/actions?query=workflow%3ACI)
[![codecov](https://codecov.io/gh/FollowTheProcess/txtar/branch/main/graph/badge.svg)](https://codecov.io/gh/FollowTheProcess/txtar)

An extended reimplementation of the [txtar] archive format 📂

## Project Description

Package txtar re-implements and extends the original [golang.org/x/tools/txtar] package, making a number of modifications and (hopefully) improvements to the package.

No modifications are made to the txtar syntax, all txtar archives produced with this package are compatible with the original.

Improvements include:

- Files stored in the archive may may be looked up by name and operated on individually
- Methods and functions are provided to help easily facilitate individual file editing
- An ergonomic API for constructing an archive, rather than simply exposing struct fields
- File names and contents are stored with all leading and trailing whitespace trimmed so that formatting the archive is easier and more consistent
- Parsing an archive from it's serialised format *can* error in the presence of a malformed document
- Parse accepts an `io.Reader` rather than a `[]byte` for greater flexibility
- Dump is provided to serialise an archive to an `io.Writer`

## Quickstart

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/FollowTheProcess/txtar"
)

func main() {
    file, err := os.Open("archive.txtar")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    archive, err := txtar.Parse(file)
    if err != nil {
        log.Fatal(err)
    }

    // Do things with the archive
    fmt.Println(archive.Comment())
    for file, contents := range archive.Files() {
        fmt.Printf("file: %s\ncontents: %s\n", string(contents))
    }
}
```

### Credits

Inspired and adapted from the original source <https://pkg.go.dev/golang.org/x/tools/txtar>, all credit to the original Go Authors. Licensed under BSD-3-Clause.

[txtar]: https://pkg.go.dev/golang.org/x/tools/txtar
