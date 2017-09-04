package vfs

import (
	"os"
	pathpkg "path"
)

func OneFile(srcFile, dstFileName string) FileSystem {
	return oneFileFileSystem{
		src: srcFile,
		dst: dstFileName,
	}
}

type oneFileFileSystem struct {
	src string
	dst string
}

func (fs oneFileFileSystem) String() string {
	return "one file"
}

func (fs oneFileFileSystem) Open(path string) (ReadSeekCloser, error) {
	if path != pathpkg.Clean("/"+fs.dst) {
		return nil, os.ErrNotExist
	}
	return os.Open(fs.src)
}

func (fs oneFileFileSystem) Lstat(path string) (os.FileInfo, error) {
	if path == "/" {
		return dirInfo("/"), nil
	}
	if path != pathpkg.Clean("/"+fs.dst) {
		return nil, os.ErrNotExist
	}
	return os.Lstat(fs.src)
}

func (fs oneFileFileSystem) Stat(path string) (os.FileInfo, error) {
	if path == "/" {
		return dirInfo("/"), nil
	}
	if path != pathpkg.Clean("/"+fs.dst) {
		return nil, os.ErrNotExist
	}
	return os.Stat(fs.src)
}

func (fs oneFileFileSystem) ReadDir(path string) ([]os.FileInfo, error) {
	if path == "/" {
		fi, err := os.Stat(fs.src)
		if err != nil {
			return nil, err
		}
		rfi := renamedFileInfo{fi, fs.dst}
		return []os.FileInfo{rfi}, nil
	}
	return nil, os.ErrNotExist
}

// renamedFile
type renamedFileInfo struct {
	os.FileInfo
	newName string
}

func (r renamedFileInfo) Name() string {
	return r.newName
}
