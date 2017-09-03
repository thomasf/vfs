package vfs

import (
	"os"
	pathpkg "path"
)

func Filter(parent FileSystem, filterFunc PathFilterFunc) FileSystem {
	return filterFileSystem{
		parent: parent,
		ff:     filterFunc,
	}
}

func Exclude(parent FileSystem, paths ...string) FileSystem {
	return filterFileSystem{
		parent: parent,
		ff: func(path string) bool {
			for _, p := range paths {
				if p == path {
					return false
				}
			}
			return true
		},
	}
}

func Include(parent FileSystem, paths ...string) FileSystem {
	return filterFileSystem{
		parent: parent,
		ff: func(path string) bool {
			for _, p := range paths {
				if hasPathPrefix(p, path) {
					return true
				}
			}
			return false
		},
	}
}

// PathFilterFunc returns true if the path should be included.
type PathFilterFunc func(path string) bool

type filterFileSystem struct {
	parent FileSystem
	ff     PathFilterFunc
}

func (fs filterFileSystem) String() string {
	return fs.parent.String()
}

func (fs filterFileSystem) Open(path string) (ReadSeekCloser, error) {
	if !fs.ff(path) {
		return nil, os.ErrNotExist
	}
	return fs.parent.Open(path)
}

func (fs filterFileSystem) Lstat(path string) (os.FileInfo, error) {
	if !fs.ff(path) {
		return nil, os.ErrNotExist
		// return nil, &os.PathError{Op: "stat", Path: path, Err: os.ErrNotExist}
	}

	return fs.parent.Lstat(path)
}

func (fs filterFileSystem) Stat(path string) (os.FileInfo, error) {
	if !fs.ff(path) {

		return nil, os.ErrNotExist
		// return nil, &os.PathError{Op: "stat", Path: path, Err: os.ErrNotExist}
	}
	return fs.parent.Stat(path)
}

func (fs filterFileSystem) ReadDir(path string) ([]os.FileInfo, error) {
	if !fs.ff(path) {
		return nil, os.ErrNotExist
	}
	dir, err := fs.parent.ReadDir(path)
	var fdir []os.FileInfo
	for _, v := range dir {
		if fs.ff(pathpkg.Join(path, v.Name())) {
			fdir = append(fdir, v)
		}
	}
	return fdir, err
}
