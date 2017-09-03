package vfs

import (
	"os"
	pathpkg "path"
	"path/filepath"
	"sort"
)

// Walk walks the file tree rooted at root, calling walkFn for each file or
// directory in the tree, including root. All errors that arise visiting files
// and directories are filtered by walkFn. The files are walked in lexical
// order, which makes the output deterministic but means that for very
// large directories Walk can be inefficient.
// Walk does not follow symbolic links.
func Walk(root string, ns NameSpace, walkFn filepath.WalkFunc) error {
	info, err := ns.Lstat(root)
	if err != nil {
		err = walkFn(root, nil, err)
	} else {
		err = walk(ns, root, info, walkFn)
	}
	if err == filepath.SkipDir {
		return nil
	}
	return err
}

// walk recursively descends path, calling w.
func walk(ns NameSpace, path string, info os.FileInfo, walkFn filepath.WalkFunc) error {
	err := walkFn(path, info, nil)
	if err != nil {
		if info.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return nil
	}

	names, err := readDirNames(ns, path)
	if err != nil {
		return walkFn(path, info, err)
	}

	for _, name := range names {
		filename := pathpkg.Join(path, name)
		fileInfo, err := ns.Lstat(filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = walk(ns, filename, fileInfo, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}
	return nil
}

// readDirNames reads the directory named by dirname and returns
// a sorted list of directory entries.
func readDirNames(ns NameSpace, dirname string) ([]string, error) {
	
	dir, err := ns.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, f := range dir {
		names = append(names, f.Name())
	}
	sort.Strings(names)
	return names, nil
}
