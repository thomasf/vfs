package vfs

import (
	"fmt"
	"os"
	pathpkg "path"
	"strings"

	"github.com/pkg/errors"
)

func Exclude(parent FileSystem, patterns ...string) FileSystem {
	return filterFileSystem{
		parent:   parent,
		patterns: patterns,
	}
}

// SafeExlude automatically adds leading slash if it doesnt exist.
func SafeExclude(parent FileSystemFunc, patterns ...string) FileSystemFunc {
	return func() (FileSystem, error) {
		par, err := parent()
		if err != nil {
			return nil,err
		}

		for i, p := range patterns {
			p = strings.TrimSpace(p)
			if !strings.HasPrefix(p, "/") {
				p = "/" + p
			}
			if strings.TrimSpace(p) == "/" {
				return nil, errors.New("emtpy pattern not allowed")
			}
			patterns[i] = p

		}
		return Exclude(par, patterns...), nil
	}
}

type filterFileSystem struct {
	parent   FileSystem
	patterns []string
}

func (fs filterFileSystem) isMatch(path string) bool {
	for _, p := range fs.patterns {
		if hasPathPrefix(path, p) {
			return false
		}
	}
	return true
}

func (fs filterFileSystem) String() string {
	return fmt.Sprintf("exclude(%s)", fs.parent.String())
}

func (fs filterFileSystem) Open(path string) (ReadSeekCloser, error) {
	if !fs.isMatch(path) {
		return nil, os.ErrNotExist
	}
	return fs.parent.Open(path)
}

func (fs filterFileSystem) Lstat(path string) (os.FileInfo, error) {
	if !fs.isMatch(path) {
		return nil, os.ErrNotExist
	}
	return fs.parent.Lstat(path)
}

func (fs filterFileSystem) Stat(path string) (os.FileInfo, error) {
	if !fs.isMatch(path) {
		return nil, os.ErrNotExist
	}
	return fs.parent.Stat(path)
}

func (fs filterFileSystem) ReadDir(path string) ([]os.FileInfo, error) {
	if !fs.isMatch(path) {
		return nil, os.ErrNotExist
	}
	dir, err := fs.parent.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var fdir []os.FileInfo
	for _, v := range dir {
		if fs.isMatch(pathpkg.Join(path, v.Name())) {
			fdir = append(fdir, v)
		}
	}
	return fdir, nil
}
