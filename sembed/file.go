package sembed

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"syscall"
	"time"
)

// http.File
//
//	type File interface {
//		io.Closer x
//		io.Reader x
//		io.Seeker x
//		Readdir(count int) ([]fs.FileInfo, error) x
//		Stat() (fs.FileInfo, error)
//	}
type File struct {
	// reader *bytes.Reader //Read, Seek
	f fs.File // need implement Seek, ReadAt if is file, ReadDir(count int) ([]fs.DirEntry, error) if is Dir
	// FileReadDir
	// fs.ReadDirFS //Open, ReadDir
	// fullPath    string
	fs.FileInfo //Name, Size, Mod, ModTime, IsDir, Sys
}

type ReadDir interface {
	// ReadDir reads the named directory
	// and returns a list of directory entries sorted by filename.
	ReadDir(count int) ([]fs.DirEntry, error)
	// ReadDir(name string) ([]fs.DirEntry, error)
}

type readDirFile struct {
	fs.FileInfo
}

func (di *readDirFile) Info() (fs.FileInfo, error) {
	return di.FileInfo, nil
}

func (di *readDirFile) Type() fs.FileMode {
	return di.FileInfo.Mode()
}

func (di *readDirFile) ModTime() time.Time {
	return time.Now()
}

// Readdir returns an empty slice of files, directory
// listing is disabled.
// func (f *FileReadDir) Readdir(count int) ([]os.FileInfo, error) {
func (f *File) Readdir(count int) (fis []os.FileInfo, err error) {
	// func (f *File) Readdir(count int) (fis []fs.DirEntry, err error) {
	if f.IsDir() {
		if rd, ok := f.f.(ReadDir); ok {
			// return rd.ReadDir(count)
			var entryDirs []fs.DirEntry
			entryDirs, err = rd.ReadDir(count)
			if err != nil {
				return nil, err
			}
			for _, dirInfo := range entryDirs {
				var fileInfo fs.FileInfo
				if fileInfo, err = dirInfo.Info(); err != nil {
					return
				}
				direntry := new(readDirFile)
				direntry = &readDirFile{
					FileInfo: fileInfo,
				}
				fis = append(fis, direntry.FileInfo)
			}
			return
		} else {
			return nil, fmt.Errorf("not implement ReadDir(name)")
		}
	} else {
		return fis, nil
	}
}

func NewFile(fs fs.File) (f *File) {
	f = new(File)
	// if len(fullPath) != 0 {
	// 	f.fullPath = fullPath[0]
	// }
	f.f = fs
	return
}

func (f *File) Close() error {
	return f.f.Close()
}

// Read reads bytes into p, returns the number of read bytes.
func (f *File) Read(p []byte) (n int, err error) {
	return f.f.Read(p)
}

// Seek seeks to the offset.
func (f *File) Seek(offset int64, whence int) (ret int64, err error) {
	if s, ok := f.f.(io.Seeker); ok {
		return s.Seek(offset, whence)
	} else {
		return 0, fmt.Errorf("not implement Seek")
	}
}

func (f *File) Stat() (fs.FileInfo, error) {
	// return f.f.Stat()
	s, e := f.f.Stat()
	if !s.IsDir() {
		osf := osF{s}
		return &osf, e
	} else {
		return f.f.Stat()
	}
}

// IsDir returns true if the file location represents a directory.
func (f *File) IsDir() bool {
	return f.FileInfo.IsDir()
}

func (f *File) Name() string {
	// if len(f.fullPath) != 0 {
	// 	return f.fullPath
	// }
	return f.FileInfo.Name()
}

func (f *File) ReadAt(p []byte, off int64) (n int, err error) {
	if ra, ok := f.f.(io.ReaderAt); ok {
		return ra.ReadAt(p, off)
	} else {
		return 0, fmt.Errorf("not implement ReaderAt")
	}
}

func (f *File) Readdirnames(count int) (names []string, err error) {
	var embedFiles []fs.FileInfo
	if embedFiles, err = f.Readdir(count); err != nil {
		return
	}
	for _, d := range embedFiles {
		// names = append(names, filepath.Join(f.fullPath, d.Name()))
		names = append(names, d.Name())

		if count > 0 && len(names) >= count {
			break
		}
	}

	// ret
	// if embedFiles, err = f.ReadDir(f.fullPath); err != nil {
	// 	return
	// }
	// for _, d := range embedFiles {
	// 	names = append(names, filepath.Join(f.fullPath, d.Name()))
	// 	if count > 0 && len(names) >= count {
	// 		break
	// 	}
	// }
	return
}

func (f *File) Sync() error                                    { return nil }
func (f *File) Truncate(size int64) error                      { return syscall.EPERM }
func (f *File) WriteString(s string) (ret int, err error)      { return 0, syscall.EPERM }
func (f *File) Write(p []byte) (n int, err error)              { return 0, syscall.EPERM }
func (f *File) WriteAt(p []byte, off int64) (n int, err error) { return 0, syscall.EPERM }
