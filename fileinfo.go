package vfs

import (
	"io"
	"os"
	pathpkg "path"
	"time"
)

func renamedFileInfo(fi os.FileInfo, name string) os.FileInfo {
	return renamedFI{fi, name}
}

func modeFileInfo(fi os.FileInfo, mode os.FileMode) os.FileInfo {
	return modeFI{fi, mode}
}

func mapFileInfo(name, contents string) os.FileInfo {
	return mapFI{name: pathpkg.Base(name), size: len(contents)}
}

func mapDirInfo(name string) os.FileInfo {
	return mapFI{name: pathpkg.Base(name), dir: true}
}

/// osPathFI implements the OSPather interface
type osPathFI struct {
	os.FileInfo
	string
}

func (o osPathFI) OSPath() string {
	return o.string
}

// renamedFileInfo wraps a os.FileInfo with a new name.
type renamedFI struct {
	os.FileInfo
	newName string
}

func (r renamedFI) Name() string {
	return r.newName
}

// modeFI wraps a os.FileInfo with a new Mode response.
type modeFI struct {
	os.FileInfo
	m os.FileMode
}

func (r modeFI) Mode() os.FileMode {
	return r.m
}

// mapFI is the map-based implementation of FileInfo.
type mapFI struct {
	name string
	size int
	dir  bool
}

func (fi mapFI) IsDir() bool        { return fi.dir }
func (fi mapFI) ModTime() time.Time { return time.Time{} }
func (fi mapFI) Mode() os.FileMode {
	if fi.IsDir() {
		return 0755 | os.ModeDir
	}
	return 0444
}
func (fi mapFI) Name() string     { return pathpkg.Base(fi.name) }
func (fi mapFI) Size() int64      { return int64(fi.size) }
func (fi mapFI) Sys() interface{} { return nil }

type nopCloser struct {
	io.ReadSeeker
}

func (nc nopCloser) Close() error { return nil }
