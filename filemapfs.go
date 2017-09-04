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
)

// FileMap returns a new FileSystem from the provided Map. The Map value
// specifies the source location of a file. Map keys should be forward
// slash-separated pathnames and not contain a leading slash.
func FileMap(m map[string]string) FileSystem {
	return filemapFS(m)
}

// filemapFS is the map based implementation of FileSystem
type filemapFS map[string]string

func (fs filemapFS) String() string {
	return fmt.Sprintf("filemap(%v)", len(fs))
}

func (fs filemapFS) Close() error { return nil }

func (fs filemapFS) Open(p string) (ReadSeekCloser, error) {
	b, ok := fs[filename(p)]
	if !ok {
		return nil, os.ErrNotExist
	}
	f, err := os.Open(b)
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if fi.IsDir() {
		f.Close()
		return nil, fmt.Errorf("Open: %s is a directory", p)
	}

	return f, nil
}

func (fs filemapFS) Lstat(p string) (os.FileInfo, error) {
	b, ok := fs[filename(p)]
	if ok {

		fi, err := os.Lstat(b)
		if err != nil {
			return nil, err
		}

		return renamedFileInfo(fi, b), nil
	}
	ents, _ := fs.ReadDir(p)
	if len(ents) > 0 {
		return mapDirInfo(p), nil
	}
	return nil, os.ErrNotExist
}

func (fs filemapFS) Stat(p string) (os.FileInfo, error) {
	return fs.Lstat(p)
}

func (fs filemapFS) ReadDir(p string) ([]os.FileInfo, error) {
	p = pathpkg.Clean(p)
	var ents []string
	fim := make(map[string]os.FileInfo) // base -> fi
	for fn, dst := range fs {
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
						var err error
						fi, err = os.Stat(dst)
						if err != nil {
							return nil, err
						}
						fi = renamedFileInfo(fi, base)
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
