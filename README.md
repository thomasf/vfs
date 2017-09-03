Simple read only virtual file system for Go

code initially copied from https://godoc.org/golang.org/x/tools/godoc/vfs

I had a look at some other VFS packages but I needed something different and
the godoc vfs was nice so I started from that.


changes from godoc version:

- removed some godoc specific code left.

- namespaces merges files inside directories as well as directory trees.

- added Walk (modified from filepath.Walk)

- added Filter vfs which wraps another vfs to include/exclude contents
  (PathFilterFunc w with convinience Include/Exclude shorthands)
  
- added OneFile vfs which mounts a single file with support for renaming it.

