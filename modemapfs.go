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
	"strings"

	"github.com/pkg/errors"
)

// ModeMap wraps a FileSystem and adds custom FileMode return values for Stat calls.
func ModeMap(fs FileSystem, m map[string]os.FileMode) FileSystem {
	return mapModeFS{fs, m}
}

func SafeModeMap(fs FileSystem, m map[string]os.FileMode) FileSystemFunc {
	return func() (FileSystem, error) {
		for path, _ := range m {
			if strings.HasPrefix(path, "/") {
				return nil, errors.Errorf("mount paths may not contain a leading '/': %s", path)
			}
		}
		return mapModeFS{fs, m}, nil
	}
}

type mapModeFS struct {
	FileSystem
	m map[string]os.FileMode
}

func (fs mapModeFS) String() string {
	return fmt.Sprintf("modemap(%v)", fs.FileSystem.String())
}

func (fs mapModeFS) Lstat(p string) (os.FileInfo, error) {
	fi, err := fs.FileSystem.Lstat(p)
	if err != nil {
		return nil, err
	}
	mode, ok := fs.m[filename(p)]
	if ok {
		return modeFileInfo(fi, mode), nil
	}
	return fi, nil
}

func (fs mapModeFS) Stat(p string) (os.FileInfo, error) {
	fi, err := fs.FileSystem.Stat(p)
	if err != nil {
		return nil, err
	}
	mode, ok := fs.m[filename(p)]
	if ok {
		return modeFileInfo(fi, mode), nil
	}
	return fi, nil
}

func (fs mapModeFS) ReadDir(p string) ([]os.FileInfo, error) {
	fis, err := fs.FileSystem.ReadDir(p)
	if err != nil {
		return fis, err
	}

	for i, fi := range fis {
		mode, ok := fs.m[filename(pathpkg.Join(p, fi.Name()))]
		if ok {
			fis[i] = modeFileInfo(fi, mode)
		}
	}
	return fis, nil
}
