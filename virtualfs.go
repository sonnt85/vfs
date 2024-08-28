package vfs

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/afero/tarfs"

	"github.com/sonnt85/gofilepath"
	"github.com/sonnt85/gosutils/sexec"
	"github.com/sonnt85/gosutils/sregexp"
	"github.com/sonnt85/gosystem"
	"github.com/spf13/afero"
)

type VFS struct {
	*AferoWrap
	// *afero.Afero
}

type AferoWrap struct {
	*afero.Afero
}

// type VFI interface {
// afero.Fs
// CreateFile(name string) (*File, error)
// AferoFsW
// }

// type AferoFsW interface {
// 	afero.Fs
// }

// type MyAferoFsW struct {
// 	AferoFsW
// }

func (sf *VFS) Create(name string) (*File, error) {
	// return sf.Fs.Create(name)
	if f, e := sf.Afero.Create(name); e == nil {
		return NewFile(f), nil
	} else {
		return nil, e
	}

}

// func (afe *VFSS) Exec(name string) (fs.File, error) {
func (vfs *VFS) Exec(rootdir, regexPath string, depth2search int, args ...interface{}) (stdob, stdeb []byte, err error) {
	var path string
	var file fs.File //[]byte
	if path, file, err = vfs.FindAndOpenFirstFileMatchRegexPathFromRoot(rootdir, regexPath, depth2search); err == nil {
		defer file.Close()
		osFilePath := ""
		switch v := file.(type) {
		case *os.File:
			osFilePath = v.Name()
		case *afero.BasePathFile:
			if off, ok := v.File.(*os.File); ok {
				osFilePath = off.Name()
			}
		case afero.File:
			if off, ok := v.(*os.File); ok {
				osFilePath = off.Name()
			}
			// default:
			// fmt.Printf("%#v", v)
		}
		if len(osFilePath) != 0 {
			return sexec.ExecCommand(osFilePath, args...)
		} else {
			return sexec.ExecBytes(file, path, args...)
		}
	}
	return
}

// func (afe *VFSS) Exec(name string) (fs.File, error) {
func (vfs *VFS) ExecFile(rootdir, filePath string, args ...interface{}) (stdob, stdeb []byte, err error) {
	var file fs.File //[]byte
	if file, err = vfs.Open(filePath); err == nil {
		defer file.Close()
		osFilePath := ""
		switch v := file.(type) {
		case *os.File:
			osFilePath = v.Name()
		case *afero.BasePathFile:
			if off, ok := v.File.(*os.File); ok {
				osFilePath = off.Name()
			}
		case afero.File:
			if off, ok := v.(*os.File); ok {
				osFilePath = off.Name()
			}
			// default:
			// fmt.Printf("%#v", v)
		}
		if len(osFilePath) != 0 {
			return sexec.ExecCommand(osFilePath, args...)
		} else {
			return sexec.ExecBytes(file, "", args...)
		}
	}
	return
}

func (afe *VFS) Open(name string) (http.File, error) {
	return afe.AferoWrap.Open(name)
	// return afe.Afero.Open(name)
}

func (afe *VFS) GetFileInfoExe(path string) fs.FileInfo {
	// f, _ := afe.Afero.Stat(path)
	// if _, ok := f.(os.FileInfo); ok {
	// 	return true
	// }
	// fmt.Printf("%#v", f)
	if s, e := afe.Stat(path); e == nil {
		return s
	} else {
		return nil
	}

	if file, e := afe.Afero.Open(path); e == nil {
		defer file.Close()
		switch v := file.(type) {
		case *afero.BasePathFile:
			if f, ok := v.File.(*tarfs.File); ok {
				if fif, err := f.Stat(); err == nil {
					return fif
				}
			}
		case afero.File:
			if v, ok := v.(*tarfs.File); ok {
				if fif, err := v.Stat(); err == nil {
					return fif
				}
			}
		}
	}
	return nil
	// afe.Stat(root string)(name string)
}

func (afe *VFS) GetRealPathOsFs(path string) (realPath string) {
	// f, _ := afe.Afero.Stat(path)
	// if _, ok := f.(os.FileInfo); ok {
	// 	return true
	// }
	// fmt.Printf("%#v", f)
	if file, e := afe.Afero.Open(path); e == nil {
		defer file.Close()
		switch v := file.(type) {
		case *os.File:
			realPath = v.Name()
		case *afero.BasePathFile:
			if f, ok := v.File.(*os.File); ok {
				realPath = f.Name()
			}
		case afero.File:
			if _, ok := v.(*os.File); ok {
				realPath = v.Name()
			}
			// default:
			// isOsFile = false
			// fmt.Printf("%#v", v)
		}

	}

	return
	// afe.Stat(root string)(name string)
}

func (afe *VFS) ReadDir(name string) (dirEntries []fs.DirEntry, err error) {
	var fif []fs.FileInfo
	if fif, err = afero.ReadDir(afe.Afero.Fs, name); err == nil {
		for _, fileInfo := range fif {
			direntry := new(readDirFile)
			direntry = &readDirFile{
				FileInfo: fileInfo,
			}
			dirEntries = append(dirEntries, direntry)
		}
		return
	} else {
		return nil, err
	}
}

func (afe *VFS) ReadDirFileInfo(name string) ([]os.FileInfo, error) {
	return afe.Afero.ReadDir(name)
	// var fif []fs.FileInfo
	// if fif, err = afero.ReadDir(afe.Afero.Fs, name); err == nil {
	// 	for _, fileInfo := range fif {
	// 		direntry := new(readDirFile)
	// 		direntry = &readDirFile{
	// 			FileInfo: fileInfo,
	// 		}
	// 		dirEntries = append(dirEntries, direntry)
	// 	}
	// 	return
	// } else {
	// 	return nil, err
	// }
}

func (vfs *VFS) ReadFile(name string) (p []byte, err error) {
	return afero.ReadFile(vfs.Fs, name)
}

// efsi *vfs.VFS, afero.Fs
func NewVFS(efsi interface{}, sub ...string) (vfs *VFS, err error) {
	// func NewVirtualFS(efs VFs, sub ...string) *VirtualFS {
	// var efs *VFs
	var baseAFs afero.Fs
	// var efi interface{}
	switch v := efsi.(type) {
	case *VFS:
		// efs = v
		baseAFs = v.Afero.Fs
	case afero.Fs:
		baseAFs = v
	default:
		return nil, fmt.Errorf("not suport type efsi")
	}
	vfs = new(VFS)
	vfs.AferoWrap = new(AferoWrap)
	if len(sub) != 0 && len(sub[0]) != 0 {
		vfs.AferoWrap.Afero = &afero.Afero{Fs: afero.NewBasePathFs(baseAFs, sub[0])}
	} else {
		vfs.AferoWrap.Afero = &afero.Afero{Fs: baseAFs}
	}
	return
}

func (vfs *AferoWrap) Open(name string) (hf http.File, err error) {
	// vfs.Fsw.Open(name)
	// afero.NewHttpFs(source afero.Fs)
	var f fs.File
	var ok bool
	f, err = vfs.Afero.Open(name)
	if err == nil {
		f.(afero.File).Seek(0, 0)
		if hf, ok = f.(http.File); ok {
			return hf, nil
		}
	}
	return nil, err

	// var file = NewFile(name)
	// var fileConten []byte

	// var fstat fs.FileInfo
	// fstat, err = vfs.Stat(name)
	// if err != nil {
	// 	return
	// }
	// file.FileInfo = fstat
	// file.shortName = name

	// file.ReadDirFS = vfs.Fsw

	// if !fstat.IsDir() {
	// 	if fileConten, err = vfs.ReadFile(name); err != nil {
	// 		return
	// 	}
	// 	file.reader = bytes.NewReader(fileConten)
	// }
	// file.fullPath = name
	// return file, nil
}

func (vfs *VFS) CreateFile(name string) (*File, error) {
	return vfs.Create(name)
}

func (vfs *VFS) OpenRDONLY(name string) (hf *File, err error) {
	if fa, e := vfs.Afero.OpenFile(name, os.O_RDONLY, os.ModePerm); e == nil {
		return NewFile(fa), nil
	} else {
		return nil, e
	}
}

func (vfs *VFS) OpenFileV(name string, flag int, perm os.FileMode) (vf *File, err error) {
	if fa, e := vfs.Afero.OpenFile(name, flag, perm); e == nil {
		return NewFile(fa), nil
	} else {
		return nil, e
	}
}

// read content first match file
func (vfs *VFS) FindAndReadFirstFileMatchRegexPathFromRoot(rootdir, regexPath string, depth2search int) (path string, bs []byte, err error) {
	for _, d := range strings.Split(rootdir, ":") {
		d = strings.TrimSpace(d)
		if matches := vfs.FindFilesMatchRegexpPathFromRoot(d, regexPath, depth2search, true, false); len(matches) != 0 {
			path = matches[0]
			if bs, err = vfs.ReadFile(path); err == nil {
				return
			}
		}
	}

	// if matches := vfs.FindFilesMatchRegexpPathFromRoot(rootdir, regexPath, depth2search, true, false); len(matches) != 0 {
	// 	path = matches[0]
	// 	if bs, err = vfs.ReadFile(path); err == nil {
	// 		return
	// 	}
	// }
	err = fmt.Errorf("can not found file with regex path %s'", regexPath)
	return
}

// read content first match file
func (vfs *VFS) FindAndOpenFirstFileMatchRegexPathFromRoot(rootdir, regexPath string, depth2search int) (path string, bs fs.File, err error) {
	for _, d := range strings.Split(rootdir, ":") {
		d = strings.TrimSpace(d)
		if matches := vfs.FindFilesMatchRegexpPathFromRoot(d, regexPath, depth2search, true, false); len(matches) != 0 {
			path = matches[0]
			if bs, err = vfs.Afero.Open(path); err == nil {
				return
			}
		}
	}
	err = fmt.Errorf("can not found file with regex path %s'", regexPath)
	return
}

// FindFilesMatchPathFromRoot searches for files that match a specified pattern starting from the root directory.
//
// Parameters:
// - rootSearch: the root directory to start the search from.
// - pattern: the pattern to match against the file paths.
// - maxdeep: the maximum depth to search for files.
// - matchfile: a boolean indicating whether to match files.
// - matchdir: a boolean indicating whether to match directories.
// - matchFunc: a function that defines the matching criteria. If matchFunc is nil, the function will exit early.
//
// Returns:
// An array of strings containing the paths of the matched files.
func (vfs *VFS) FindFilesMatchPathFromRoot(rootSearch, pattern string, maxdeep int, matchfile, matchdir bool, matchFunc func(pattern, relpath string) bool) (matches []string) {
	matches = make([]string, 0)
	if matchFunc == nil {
		return
	}
	// if len(rootSearch) == 0 || rootSearch == "/" || rootSearch == "." || rootSearch == "./" {
	// 	rootSearch = vfs.rootDir
	// }
	// rootSearch := gofilepath.FromSlash(root1)
	if finfo, err := vfs.Stat(rootSearch); err == nil {
		if !finfo.IsDir() { //is file
			if matchFunc(pattern, rootSearch) {
				matches = []string{rootSearch}
			}
			return
		}
	}
	// pattern = gofilepath.ToSlash(pattern)
	var relpath string
	var deep int
	if nil != vfs.WalkDir(rootSearch, func(path string, d fs.DirEntry, err error) error {
		if err != nil { //signaling that Walk will not walk into this directory.
			// return err
			return nil
		}
		relpath, err = gofilepath.RelSmart(rootSearch, path)
		if err != nil {
			return nil
		}
		if maxdeep > -1 {
			deep = gofilepath.CountPathSeparator(relpath)
			if deep > maxdeep {
				if d.IsDir() {
					return fs.SkipDir
				} else {
					return nil
				}
			}
		}
		if (d.IsDir() && matchdir) || (!d.IsDir() && matchfile) {
			if matchFunc(pattern, relpath) {
				matches = append(matches, path)
			}
		}
		return nil
	}) {
		return nil
	}
	return matches
}

// maxdeep: 0 ->
// func (vfs *VirtualFS) _FindFilesMatchPathFromRoot(root, pattern string, maxdeep int, matchfile, matchdir bool, matchFunc func(pattern, relpath string) bool) (matches []string) {
// 	return gofilepath.FindFilesMatchPathFromRoot(root, pattern, maxdeep, matchfile, matchdir, matchFunc, vfs.WalkDir)
// }

func (vfs *VFS) FindFilesMatchRegexpPathFromRoot(root, pattern string, maxdeep int, matchfile, matchdir bool) (matches []string) {
	matchFunc := func(pattern, relpath string) bool {
		return sregexp.New(pattern).MatchString(relpath)
	}
	return vfs.FindFilesMatchPathFromRoot(root, pattern, maxdeep, matchfile, matchdir, matchFunc)
}

func (vfs *VFS) FindFilesMatchRegexpName(root, pattern string, maxdeep int, matchfile, matchdir bool) (matches []string) {
	matchFunc := func(pattern, relpath string) bool {
		return sregexp.New(pattern).MatchString(gofilepath.Base(relpath))
	}
	return vfs.FindFilesMatchPathFromRoot(root, pattern, maxdeep, matchfile, matchdir, matchFunc)
}

// FindFilesMatchName finds files by matching the provided pattern in the specified root directory up to a certain depth.
//
// Parameters:
// - root: the root directory to start searching from.
// - pattern: the pattern to match against the file names.
// - maxdeep: the maximum depth to search for files.
// - matchfile: a boolean indicating whether to match files.
// - matchdir: a boolean indicating whether to match directories.
//
// Returns:
// An array of strings containing the paths of the matched files.
func (vfs *VFS) FindFilesMatchName(root, pattern string, maxdeep int, matchfile, matchdir bool) (matches []string) {
	matchFunc := func(pattern, relpath string) bool {
		if match, err := filepath.Match(pattern, gofilepath.Base(relpath)); err == nil && match {
			return true
		}
		return false
	}
	return vfs.FindFilesMatchPathFromRoot(root, pattern, maxdeep, matchfile, matchdir, matchFunc)
}

func (vfs *VFS) Stat(root string) (finfo fs.FileInfo, err error) {
	// if len(root) == 0 || root == "/" || root == "./" {
	// 	root = "."
	// }
	// vfs.Afero.Stat("root")
	return vfs.Afero.Stat(root)
	file, err := vfs.Afero.Open(root)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return file.Stat()
}

func (vfs *VFS) WalkDir(root string, fn WalkDirFunc) (err error) {
	info, err := vfs.Stat(root)
	if err != nil {
		err = fn(root, nil, err)
	} else {
		err = vfs.walkDir(root, &statDirEntry{info}, fn)
	}
	if err == fs.SkipDir {
		return nil
	}
	return err
}

// walkDir recursively descends path, calling walkDirFn.
func (vfs *VFS) walkDir(pathdir string, d fs.DirEntry, walkDirFn WalkDirFunc) error {
	if err := walkDirFn(pathdir, d, nil); err != nil || !d.IsDir() {
		if err == fs.SkipDir && d.IsDir() {
			// Successfully skipped directory.
			err = nil
		}
		return err
	}
	dirs, err := vfs.ReadDir(pathdir)
	if err != nil {
		// Second call, to report ReadDir error.
		err = walkDirFn(pathdir, d, err)
		if err != nil {
			return err
		}
	}

	for _, d1 := range dirs {
		path1 := path.Join(pathdir, d1.Name())
		if err := vfs.walkDir(path1, d1, walkDirFn); err != nil {
			if err == fs.SkipDir {
				break
			}
			return err
		}
	}
	return nil
}

func writeFile(name string, r io.Reader, perm os.FileMode) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

// copy file or directory from vfs  to fs dirName
func (vfs *VFS) Copy(toDirPath, fromFshPath string, mods ...interface{}) (err error) {
	// defer func() {
	// 	if err != nil {
	// 		//cleanup function
	// 	}
	// }()
	// fromFshPath = vfs.Getfullpath(fromFshPath)
	cr := &CopyRecursive{IsVerbose: true,
		IgnErr:   false,
		ReadFile: vfs.ReadFile,
		Mkdir: func(srcPath string, srcFileInfo fs.FileMode) error {
			if gosystem.PathIsExist(srcPath) {
				os.Chmod(srcPath, 0755)
				// os.Chmod(srcPath, srcFileInfo)
				return nil
			}
			return os.Mkdir(srcPath, srcFileInfo)
		},
		Writer: writeFile,
		Stat:   vfs.Stat,
		Open:   vfs.OpenRDONLY,
	}
	// if len(mods) != 0 {
	// 	cr.allFileMode = mods[0]
	// }
	toDirPath = gofilepath.FromSlash(toDirPath)
	// if !gofilepath.IsAbs(toDirPath) {
	// 	toDirPath, err = gofilepath.Abs(toDirPath)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return cr.Copy(toDirPath, fromFshPath, mods...)
}
