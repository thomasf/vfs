// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package vfs defines types for abstract file system access and provides an
// implementation accessing the file system of the underlying OS.
package vfs // import "github.com/thomasf/vfs"

import (
	"io"
	"io/ioutil"
	"os"
)

// The FileSystem interface specifies the methods godoc is using
// to access the file system for which it serves documentation.
type FileSystem interface {
	Opener
	Lstat(path string) (os.FileInfo, error)
	Stat(path string) (os.FileInfo, error)
	ReadDir(path string) ([]os.FileInfo, error)
	String() string
}

// FileSystemFunc returns a FileSystem or an error if it's not configured
// correctly. Functions that returns FileSystemFuncs should verify that the
// underlying resources exists and if possible automatically fix bad
// configurations.
type FileSystemFunc func() (FileSystem, error)

// Wrapper which returns fs and no error
func Safe(fs FileSystem) FileSystemFunc {
	return func() (FileSystem, error) {
		return fs, nil
	}
}

// MustSafe
func MustSafe(f FileSystemFunc) FileSystem {
	fs, err := f()
	if err != nil {
		panic(err)
	}
	return fs
}

// OSPather contains the full path to files on vfs FileInfo instances which
// maps to an os.File path. Be careful with using OSPath results for
// directories when multiple filesystems are mounted to the same path in the
// namespace since the path returned by OSPath only leads to the first matching
// vfs.
type OSPather interface {
	OSPath() string
}

// Opener is a minimal virtual filesystem that can only open regular files.
type Opener interface {
	Open(name string) (ReadSeekCloser, error)
}

// A ReadSeekCloser can Read, Seek, and Close.
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// ReadFile reads the file named by path from fs and returns the contents.
func ReadFile(fs Opener, path string) ([]byte, error) {
	rc, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return ioutil.ReadAll(rc)
}
