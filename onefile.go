package vfs

import (
	"os"
	pathpkg "path"

	"github.com/pkg/errors"
)

// OneFile contains a link to a single OS file at the root of the VFS. The
// first argument is the full path to the local file, the second argument must
// be a single file name without any directories or leading slash.
func OneFile(path, newname string) FileSystem {
	return oneFileFileSystem{
		path: path,
		name: newname,
	}
}

func SafeOneFile(path, newname string) FileSystemFunc {
	return func() (FileSystem, error) {
		// todo , what is rules for newname?
		// if strings.HasPrefix(new, "/") {
		// 	return nil, errors.Errorf("mount paths may not contain a leading '/': %s", new)
		// }
		fi, err := os.Stat(path)
		if err != nil {
			return nil, errors.Wrapf(err, "%s is not a readable path", path)
		}
		if fi.IsDir() {
			return nil, errors.Wrapf(err, "%s is a directory, not a file", path)
		}
		return oneFileFileSystem{
			path: path,
			name: newname,
		}, nil

	}
}

type oneFileFileSystem struct {
	path string
	name string
}

func (fs oneFileFileSystem) String() string {
	return "onefile(" + fs.path + ":" + fs.name + ")"
}

func (fs oneFileFileSystem) Open(path string) (ReadSeekCloser, error) {
	if path != pathpkg.Clean("/"+fs.name) {
		return nil, os.ErrNotExist
	}
	return os.Open(fs.path)
}

func (fs oneFileFileSystem) Lstat(path string) (os.FileInfo, error) {
	if path == "/" {
		return dirInfo("/"), nil
	}
	if path != pathpkg.Clean("/"+fs.name) {
		return nil, os.ErrNotExist
	}
	fi, err := os.Lstat(fs.path)
	return osPathFI{fi, fs.path}, err
}

func (fs oneFileFileSystem) Stat(path string) (os.FileInfo, error) {
	if path == "/" {
		return dirInfo("/"), nil
	}
	if path != pathpkg.Clean("/"+fs.name) {
		return nil, os.ErrNotExist
	}
	fi, err := os.Stat(fs.path)
	return osPathFI{fi, fs.path}, err
}

func (fs oneFileFileSystem) ReadDir(path string) ([]os.FileInfo, error) {
	if path == "/" {
		fi, err := os.Stat(fs.path)
		if err != nil {
			return nil, err
		}
		rfi := renamedFileInfo(fi, fs.name)
		return []os.FileInfo{rfi}, nil
	}
	return nil, os.ErrNotExist
}
