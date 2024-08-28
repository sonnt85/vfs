package sembed

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/afero"
)

type EFs struct {
	// *HttpSystemFS
	*embed.FS
}

func (efs *EFs) Create(name string) (afero.File, error) {
	return nil, syscall.EPERM
}

func (efs *EFs) Mkdir(name string, perm os.FileMode) error { return syscall.EPERM }

func (efs *EFs) MkdirAll(path string, perm os.FileMode) error { return syscall.EPERM }

func (efs *EFs) Open(name string) (fi afero.File, e error) {
	// name = removePrefix(name, "/", "./", ".")
	var ffs fs.File
	// var fstat fs.FileInfo
	ffs, e = efs.FS.Open(name)
	if e != nil {
		return
	}
	defer ffs.Close()
	var file = NewFile(ffs)

	// fstat, e = efs.Stat(name)
	var fstat fs.FileInfo
	if fstat, e = ffs.Stat(); e == nil {
		file.FileInfo = fstat
	} else {
		return
	}

	// file.ReadDirFS = efs.FS
	return file, nil
}

func (efs *EFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag != os.O_RDONLY {
		return nil, syscall.EPERM
	}
	return efs.Open(name)
}

func (efs *EFs) Remove(name string) error { return syscall.EPERM }

func (efs *EFs) RemoveAll(path string) error { return syscall.EPERM }

func (efs *EFs) Rename(oldname, newname string) error { return syscall.EPERM }

func removePrefix(s string, prefixes ...string) string {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return strings.TrimPrefix(s, prefix)
		}
	}
	return s
}

type osF struct {
	os.FileInfo
}

func (sf *osF) Mode() fs.FileMode {
	return sf.FileInfo.Mode() | 0111
}

func (efs *EFs) Stat(name string) (os.FileInfo, error) {
	name = removePrefix(name, "/", "./", ".")
	file, err := efs.FS.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	s, e := file.Stat()
	if !s.IsDir() && strings.HasSuffix(filepath.Base(name), "bin") {
		osf := osF{s}
		return &osf, e
	} else {
		return s, e
	}
}

func (fs *EFs) Name() string { return "embedfs" }

func (fs *EFs) Chmod(name string, mode os.FileMode) error { return syscall.EPERM }

func (fs *EFs) Chown(name string, uid, gid int) error { return syscall.EPERM }

func (fs *EFs) Chtimes(name string, atime time.Time, mtime time.Time) error { return syscall.EPERM }
