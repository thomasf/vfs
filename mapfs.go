// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package mapfs file provides an implementation of the FileSystem
// interface based on the contents of a map[string]string.
package vfs // import "github.com/thomasf/vfs"

import (
	"fmt"
	"os"
	pathpkg "path"
	"sort"
	"strings"
)

// Map returns a new FileSystem from the provided map. Map keys should be
// forward slash-separated pathnames and not contain a leading slash. The Map
// value string contents is returned by Open.
func Map(m map[string]string) FileSystem {
	return mapFS(m)
}

func SafeMap(m map[string]string) FileSystemFunc {
	return func() (FileSystem, error) {

		for path, data := range m {
			if strings.HasPrefix(path, "/") {
				path = strings.TrimLeft(path, "/")
			}
			m[path] = data

		}
		return mapFS(m), nil
	}
}

// mapFS is the map based implementation of FileSystem
type mapFS map[string]string

func (fs mapFS) String() string {
	return fmt.Sprintf("filemap(%v)", len(fs))
}

func (fs mapFS) Close() error { return nil }

func filename(p string) string {
	return strings.TrimPrefix(p, "/")
}

func (fs mapFS) Open(p string) (ReadSeekCloser, error) {
	b, ok := fs[filename(p)]
	if !ok {
		return nil, os.ErrNotExist
	}
	return nopCloser{strings.NewReader(b)}, nil
}

func (fs mapFS) Lstat(p string) (os.FileInfo, error) {
	b, ok := fs[filename(p)]
	if ok {
		return mapFileInfo(p, b), nil
	}
	ents, _ := fs.ReadDir(p)
	if len(ents) > 0 {
		return mapDirInfo(p), nil
	}
	return nil, os.ErrNotExist
}

func (fs mapFS) Stat(p string) (os.FileInfo, error) {
	return fs.Lstat(p)
}

// slashdir returns path.Dir(p), but special-cases paths not beginning
// with a slash to be in the root.
func slashdir(p string) string {
	d := pathpkg.Dir(p)
	if d == "." {
		return "/"
	}
	if strings.HasPrefix(p, "/") {
		return d
	}
	return "/" + d
}

func (fs mapFS) ReadDir(p string) ([]os.FileInfo, error) {
	p = pathpkg.Clean(p)
	var ents []string
	fim := make(map[string]os.FileInfo) // base -> fi
	for fn, b := range fs {
		dir := slashdir(fn)
		isFile := true
		var lastBase string
		for {
			if dir == p {
				base := lastBase
				if isFile {
					base = pathpkg.Base(fn)
				}
				if fim[base] == nil {
					var fi os.FileInfo
					if isFile {
						fi = mapFileInfo(fn, b)
					} else {
						fi = mapDirInfo(base)
					}
					ents = append(ents, base)
					fim[base] = fi
				}
			}
			if dir == "/" {
				break
			} else {
				isFile = false
				lastBase = pathpkg.Base(dir)
				dir = pathpkg.Dir(dir)
			}
		}
	}
	if len(ents) == 0 {
		return nil, os.ErrNotExist
	}

	sort.Strings(ents)
	var list []os.FileInfo
	for _, dir := range ents {
		list = append(list, fim[dir])
	}
	return list, nil
}
