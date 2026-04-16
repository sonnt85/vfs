# vfs

[![Go Reference](https://pkg.go.dev/badge/github.com/sonnt85/vfs.svg)](https://pkg.go.dev/github.com/sonnt85/vfs)

Virtual filesystem library for Go — unified API over OS files, embedded FS, zip/tar archives, and overlay filesystems, with HTTP serving, file search, template rendering, and sync.

## Installation

```bash
go get github.com/sonnt85/vfs
```

## Features

- Wraps [afero](https://github.com/spf13/afero) with a richer `VFS` type
- Create a VFS from: OS directory, `embed.FS`, zip file/reader, tar file/reader, or any `afero.Fs`
- Overlay (copy-on-write) filesystem: layer a writable FS over a read-only base
- Implements `http.FileSystem` for serving embedded or archive content
- Search files by regex path or name pattern with depth control
- Execute files found inside a VFS (supports in-memory binary execution)
- Copy files/directories within or between VFS instances with optional template rendering
- Directory sync (`Sync`, `SyncTo`) — rsync-style directory mirroring
- `WalkDir` for recursive traversal
- File locking and create/open helpers (`Create`, `CreateFile`, `OpenRDONLY`, `OpenFileV`)

## Usage

```go
import "github.com/sonnt85/vfs"

// OS-backed VFS rooted at a directory
osVfs := vfs.NewOsFs("/var/data")

// Embedded FS (Go 1.16+ embed)
//go:embed assets
var assets embed.FS
embVfs, _ := vfs.NewEmbedHttpSystemFS(&assets, "assets")

// Zip archive
zipVfs, _ := vfs.NewZipFs("/path/to/archive.zip")

// Tar archive
tarVfs, _ := vfs.NewTarFs("/path/to/archive.tar.gz")

// Overlay: write to memory, read from zip
overlay := vfs.NewOverlayFs(zipVfs, vfs.NewOsFs("/tmp/overlay"))

// Serve embedded assets over HTTP
http.Handle("/", embVfs)

// Read a file
data, _ := embVfs.ReadFile("index.html")

// Search by regex
matches := embVfs.FindFilesMatchRegexpName("/", `\.html$`, 5, true, false)

// Execute a binary found inside the VFS
stdout, stderr, err := embVfs.Exec("/", `mybin`, 3, "--flag")

// Sync two directories
vfs.Sync("/dst/dir", "/src/dir")
```

## API

### Types

- `type VFS struct` — main virtual filesystem type (embeds `*AferoWrap`)
- `type AferoWrap struct` — wraps `*afero.Afero` and implements `http.FileSystem`
- `type File struct` — virtual file (wraps `afero.File`) with `Readdir` and `Stat`
- `type FileInfo interface` — shared interface for `os.FileInfo` and `fs.DirEntry` (`Name()`, `IsDir()`)
- `type WalkDirFunc` — alias for `fs.WalkDirFunc`

### Constructors
- `NewOsFs(dir ...string) *VFS` — OS filesystem, optionally rooted at dir
- `NewEmbedHttpSystemFS(efs *embed.FS, rootDir string, sub ...string) (*VFS, error)` — from embedded FS
- `NewEFs(efs *embed.FS, rootDir string, sub ...string) (*VFS, error)` — alias
- `NewZipFs(r interface{}, password ...[]byte) (*VFS, error)` — from zip path or `*zip.Reader`
- `NewTarFs(r interface{}, password ...[]byte) (*VFS, error)` — from tar path or `*tar.Reader`
- `NewArchiveFs(r interface{}, password ...[]byte) (*VFS, error)` — auto-detect zip or tar
- `NewOverlayFs(base, layer interface{}) *VFS` — copy-on-write overlay
- `NewVFSFromAFS(afs afero.Fs, sub ...string) *VFS` — wrap any `afero.Fs`
- `NewVFS(efsi interface{}, sub ...string) (*VFS, error)` — wrap `afero.Fs` or `*VFS`

### File Operations (VFS methods)
- `Create(name string) (*File, error)` — create new file
- `CreateFile(name string) (*File, error)` — create file (alias)
- `OpenRDONLY(name string) (*File, error)` — open for reading only
- `OpenFileV(name string, flag int, perm os.FileMode) (*File, error)` — open with flags
- `ReadFile(name string) ([]byte, error)` — read entire file
- `ReadDir(name string) ([]fs.DirEntry, error)` — list directory as `DirEntry` slice
- `ReadDirFileInfo(name string) ([]os.FileInfo, error)` — list directory as `os.FileInfo` slice
- `GetFileInfoExe(path string) fs.FileInfo` — return `FileInfo` for an executable inside the VFS
- `Stat(path string) (fs.FileInfo, error)` — file info
- `Open(name string) (http.File, error)` — open as `http.File`

### Search
- `FindFilesMatchPathFromRoot(root, pattern string, maxdeep int, matchfile, matchdir bool, matchFunc func(pattern, relpath string) bool) []string` — generic file search with custom match function
- `FindFilesMatchRegexpPathFromRoot(root, pattern string, maxdeep int, matchfile, matchdir bool) []string` — regex on full path
- `FindFilesMatchRegexpName(root, pattern string, maxdeep int, matchfile, matchdir bool) []string` — regex on filename only
- `FindFilesMatchName(root, pattern string, maxdeep int, matchfile, matchdir bool) []string` — glob on filename
- `FindAndReadFirstFileMatchRegexPathFromRoot(rootdir, regex string, depth int) (path string, data []byte, err error)` — read first match
- `FindAndOpenFirstFileMatchRegexPathFromRoot(rootdir, regex string, depth int) (path string, f fs.File, err error)` — open first match

### Execution
- `Exec(rootdir, regexPath string, depth int, args ...)` — find and execute binary
- `ExecFile(rootdir, filePath string, args ...)` — execute specific file path

### Copy & Sync
- `(*VFS).Copy(toDirPath, fromFsPath string, mods ...interface{}) error` — copy with optional template rendering
- `ParseTemplate(r io.Reader, data interface{}, w io.Writer) error` — render Go template from reader
- `Sync(dst, src string) error` — sync src directory to dst
- `SyncTo(to string, srcs ...string) error` — sync multiple sources to destination
- `NewSyncer() *Syncer` — create a reusable syncer

### Traversal
- `(*VFS).WalkDir(root string, fn WalkDirFunc) error` — walk with `fs.DirEntry`
- `(*VFS).GetRealPathOsFs(path string) string` — resolve real OS path (for BasePathFs)

## Author

**sonnt85** — [thanhson.rf@gmail.com](mailto:thanhson.rf@gmail.com)

## License

MIT License - see [LICENSE](LICENSE) for details.
