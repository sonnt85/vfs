package vfs

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/sonnt85/gofilepath"
	"github.com/sonnt85/gosutils/sregexp"
	"github.com/sonnt85/gosystem"
	"github.com/spf13/afero"
)

type VFS struct {
	*VFSS
}

type VFSS struct {
	VFI
}

type VFI interface {
	afero.Fs
	// CreateFile(name string) (*File, error)
	// AferoFsW
}

// type AferoFsW interface {
// 	afero.Fs
// }

// type MyAferoFsW struct {
// 	AferoFsW
// }

func (sf *VFSS) Create(name string) (*File, error) {
	// return sf.Fs.Create(name)
	if f, e := sf.VFI.Create(name); e == nil {
		return NewFile(f), nil
	} else {
		return nil, e
	}

}

// func (afe *VFSS) Exec(name string) (fs.File, error) {
func (afe *VFSS) Exec(regexPath string, depth2search int, args ...string) (stdoe string, err error) {
	// var path string
	// var byteprog []byte
	// if path, byteprog, err = afe.FindAndReadFirstFileMatchRegexPath(regexPath, depth2search); err == nil {
	// 	userLogDir := filepath.Join(Gvar.Config.Root_logs, Gvar.UserName)
	// 	if !gosystem.PathIsExist(userLogDir) {
	// 		os.MkdirAll(userLogDir, 0755)
	// 	}
	// 	ext := strings.ToLower(filepath.Ext(path))
	// 	fileBinPathNoExt := strings.TrimSuffix(path, ext)
	// 	logFilePath := filepath.Join(userLogDir, filepath.Base(fileBinPathNoExt)+".log")
	// 	Gvar.processLogFiles = append(Gvar.processLogFiles, logFilePath)
	// 	perm := fs.FileMode(0660)
	// 	if strings.Contains(fileBinPathNoExt, "-xdesktop") {
	// 		perm = 0666
	// 	}
	// 	var logFile *os.File
	// 	logFile, err = os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	// 	os.Chmod(logFilePath, perm)
	// 	if err != nil {
	// 		slogrus.ErrorS("can not create logfile: ", err)
	// 		return
	// 	}
	// 	defer logFile.Close()
	// 	start := time.Now()

	// 	var stdob, stdeb []byte
	// 	if len(args) == 0 {
	// 		args = make([]string, 0)
	// 	}
	// 	slogrus.InfofS("Running job %s [%d] [ blocking ] [logfile %s] ...", path, os.Getegid(), logFilePath)
	// 	stdob, stdeb, err = sexec.ExecBytes(byteprog, gofilepath.Base(path), sexec.CombineArgsAndStreams(args, logFile, logFile)...)
	// 	stdoe = string(stdob) + "\n" + string(stdeb)
	// 	slogrus.Infof("Finish job %s [blocking] [Period %s]", path, time.Since(start))
	// }
	// return
	return
}

func (afe *VFSS) Open(name string) (fs.File, error) {
	return afe.VFI.Open(name)
}

func (afe *VFSS) ReadDir(name string) (dirEntries []fs.DirEntry, err error) {
	var fif []fs.FileInfo
	if fif, err = afero.ReadDir(afe.VFI, name); err == nil {
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

func (fsh *VFS) ReadFile(name string) (p []byte, err error) {
	return afero.ReadFile(fsh.VFI, name)
}

func (esub *VFS) ReadDir(name string) (dirEntries []fs.DirEntry, err error) {
	var fif []fs.FileInfo
	if fif, err = afero.ReadDir(esub.VFI, name); err == nil {
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

// efsi *vfs.VFS, afero.Fs
func NewVFS(efsi interface{}, sub ...string) (*VFS, error) {
	// func NewVirtualFS(efs VFs, sub ...string) *VirtualFS {
	// var efs *VFs
	var baseAFs afero.Fs
	// var efi interface{}
	switch v := efsi.(type) {
	case *VFS:
		// efs = v
		baseAFs = v.VFI
	case afero.Fs:
		baseAFs = v
	default:
		return nil, fmt.Errorf("not suport type efsi")
	}

	if len(sub) != 0 && len(sub[0]) != 0 {
		return &VFS{
			VFSS: &VFSS{
				afero.NewBasePathFs(baseAFs, sub[0]),
			},
		}, nil
	} else {
		return &VFS{
			VFSS: &VFSS{
				baseAFs,
			},
		}, nil
	}
}

func (fsh *VFS) Open(name string) (hf http.File, err error) {
	// fsh.Fsw.Open(name)
	// afero.NewHttpFs(source afero.Fs)
	var f fs.File
	var ok bool
	f, err = fsh.VFSS.VFI.Open(name)
	if err == nil {
		if hf, ok = f.(http.File); ok {
			return hf, nil
		}
	}
	return nil, err

	// var file = NewFile(name)
	// var fileConten []byte

	// var fstat fs.FileInfo
	// fstat, err = fsh.Stat(name)
	// if err != nil {
	// 	return
	// }
	// file.FileInfo = fstat
	// file.shortName = name

	// file.ReadDirFS = fsh.Fsw

	// if !fstat.IsDir() {
	// 	if fileConten, err = fsh.ReadFile(name); err != nil {
	// 		return
	// 	}
	// 	file.reader = bytes.NewReader(fileConten)
	// }
	// file.fullPath = name
	// return file, nil
}

func (fsh *VFS) CreateFile(name string) (*File, error) {
	return fsh.Create(name)
}

func (fsh *VFS) OpenRDONLY(name string) (hf *File, err error) {
	if fa, e := fsh.VFI.OpenFile(name, os.O_RDONLY, os.ModePerm); e == nil {
		return NewFile(fa), nil
	} else {
		return nil, e
	}
}

func (fsh *VFS) OpenFileV(name string, flag int, perm os.FileMode) (hf *File, err error) {
	if fa, e := fsh.VFI.OpenFile(name, flag, perm); e == nil {
		return NewFile(fa), nil
	} else {
		return nil, e
	}
}

// read content first match file
func (fsh *VFS) FindAndReadFirstFileMatchRegexPath(regexPath string, depth2search int) (path string, bs []byte, err error) {
	if matches := fsh.FindFilesMatchRegexpPathFromRoot("", regexPath, depth2search, true, false); len(matches) != 0 {
		path = matches[0]
		if bs, err = fsh.ReadFile(path); err == nil {
			return
		}
	}
	err = fmt.Errorf("can not found file with regex path %s'", regexPath)
	return
}

func (fsh *VFS) FindFilesMatchPathFromRoot(rootSearch, pattern string, maxdeep int, matchfile, matchdir bool, matchFunc func(pattern, relpath string) bool) (matches []string) {
	matches = make([]string, 0)
	if matchFunc == nil {
		return
	}
	// if len(rootSearch) == 0 || rootSearch == "/" || rootSearch == "." || rootSearch == "./" {
	// 	rootSearch = fsh.rootDir
	// }
	// rootSearch := gofilepath.FromSlash(root1)
	if finfo, err := fsh.Stat(rootSearch); err == nil {
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
	if nil != fsh.WalkDir(rootSearch, func(path string, d fs.DirEntry, err error) error {
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
// func (fsh *VirtualFS) _FindFilesMatchPathFromRoot(root, pattern string, maxdeep int, matchfile, matchdir bool, matchFunc func(pattern, relpath string) bool) (matches []string) {
// 	return gofilepath.FindFilesMatchPathFromRoot(root, pattern, maxdeep, matchfile, matchdir, matchFunc, fsh.WalkDir)
// }

func (fsh *VFS) FindFilesMatchRegexpPathFromRoot(root, pattern string, maxdeep int, matchfile, matchdir bool) (matches []string) {
	matchFunc := func(pattern, relpath string) bool {
		return sregexp.New(pattern).MatchString(relpath)
	}
	return fsh.FindFilesMatchPathFromRoot(root, pattern, maxdeep, matchfile, matchdir, matchFunc)
}

func (fsh *VFS) FindFilesMatchRegexpName(root, pattern string, maxdeep int, matchfile, matchdir bool) (matches []string) {
	matchFunc := func(pattern, relpath string) bool {
		return sregexp.New(pattern).MatchString(gofilepath.Base(relpath))
	}
	return fsh.FindFilesMatchPathFromRoot(root, pattern, maxdeep, matchfile, matchdir, matchFunc)
}

func (fsh *VFS) FindFilesMatchName(root, pattern string, maxdeep int, matchfile, matchdir bool) (matches []string) {
	matchFunc := func(pattern, relpath string) bool {
		if match, err := filepath.Match(pattern, gofilepath.Base(relpath)); err == nil && match {
			return true
		}
		return false
	}
	return fsh.FindFilesMatchPathFromRoot(root, pattern, maxdeep, matchfile, matchdir, matchFunc)
}

func (fsh *VFS) Stat(root string) (finfo fs.FileInfo, err error) {
	// if len(root) == 0 || root == "/" || root == "./" {
	// 	root = "."
	// }
	file, err := fsh.VFSS.VFI.Open(root)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return file.Stat()
}

func (fsh *VFS) WalkDir(root string, fn WalkDirFunc) (err error) {
	info, err := fsh.Stat(root)
	if err != nil {
		err = fn(root, nil, err)
	} else {
		err = fsh.walkDir(root, &statDirEntry{info}, fn)
	}
	if err == fs.SkipDir {
		return nil
	}
	return err
}

// walkDir recursively descends path, calling walkDirFn.
func (fsh *VFS) walkDir(pathdir string, d fs.DirEntry, walkDirFn WalkDirFunc) error {
	if err := walkDirFn(pathdir, d, nil); err != nil || !d.IsDir() {
		if err == fs.SkipDir && d.IsDir() {
			// Successfully skipped directory.
			err = nil
		}
		return err
	}
	dirs, err := fsh.ReadDir(pathdir)
	if err != nil {
		// Second call, to report ReadDir error.
		err = walkDirFn(pathdir, d, err)
		if err != nil {
			return err
		}
	}

	for _, d1 := range dirs {
		path1 := path.Join(pathdir, d1.Name())
		if err := fsh.walkDir(path1, d1, walkDirFn); err != nil {
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

// copy file or directory from fsh  to fs dirName
func (fsh *VFS) Copy(toDirPath, fromFshPath string, mods ...fs.FileMode) (err error) {
	// defer func() {
	// 	if err != nil {
	// 		//cleanup function
	// 	}
	// }()
	// fromFshPath = fsh.Getfullpath(fromFshPath)
	cr := &CopyRecursive{IsVerbose: true,
		IgnErr:   false,
		ReadFile: fsh.ReadFile,
		Mkdir: func(srcPath string, srcFileInfo fs.FileMode) error {
			if gosystem.PathIsExist(srcPath) {
				os.Chmod(srcPath, 0755)
				// os.Chmod(srcPath, srcFileInfo)
				return nil
			}
			return os.Mkdir(srcPath, srcFileInfo)
		},
		Writer: writeFile,
		Stat:   fsh.Stat,
		Open:   fsh.OpenRDONLY,
	}
	toDirPath = gofilepath.FromSlash(toDirPath)
	// if !gofilepath.IsAbs(toDirPath) {
	// 	toDirPath, err = gofilepath.Abs(toDirPath)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return cr.Copy(toDirPath, fromFshPath, mods...)
}
