Simple read only virtual file system for Go

code initially copied from https://godoc.org/golang.org/x/tools/godoc/vfs

I had a look at some other VFS packages but I needed something different and
the godoc vfs was nice so I started from that.


changes from godoc version:

- removed godoc specific code.

- namespaces merges files inside directories as well as directories.

- added Walk (modified from filepath.Walk)

- added Exclude vfs which wraps another vfs to exclude by path prefix.
  
- added OneFile vfs which mounts a single file with support for renaming it.

