package vfs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"

	"github.com/sonnt85/gofilepath"
)

// type WriteFun func(name string, data []byte, perm os.FileMode) error
type WriteFun func(name string, r io.Reader, perm os.FileMode) error

type MkDirFun func(srcPath string, srcFileInfo fs.FileMode) error

type CopyRecursive struct {
	// IsRecursive bool
	// http.FileSystem
	IsVerbose bool

	// fs.ReadFileFS
	// fs.StatFS
	Open func(name string) (*File, error)

	// Open     func(name string) (http.File, error)
	Stat     func(root string) (finfo fs.FileInfo, err error)
	ReadFile func(name string) ([]byte, error)
	Writer   WriteFun
	Mkdir    MkDirFun

	IgnErr                        bool
	srcPath                       string
	dstPath                       string
	srcPathSeparator              string
	dstPathSeparator              string
	matchFilNameRegexpForTemplate string
	isTemplateMode                bool
	envs                          []string
	dataTmpl                      interface{}
	allFileMode                   fs.FileMode
}

func (cr *CopyRecursive) mkdir(srcPath string, srcFileInfo fs.FileMode) error {
	// return cr.Mkdir(srcPath, 0755)
	return cr.Mkdir(srcPath, srcFileInfo.Perm()|0200)

}

func (cr *CopyRecursive) processDir(srcFilePath string, srcFileInfo os.FileInfo) (err error) {
	var relpath string
	relpath, err = gofilepath.Rel(cr.srcPath, srcFilePath)
	if err != nil {
		return
	}
	newdir := gofilepath.JointSmart(cr.dstPathSeparator, cr.dstPath, relpath)
	err = cr.mkdir(newdir, srcFileInfo.Mode())
	if err != nil {
		return err
	}
	dir, err := cr.Open(srcFilePath)
	if err != nil {
		return err
	}
	fis, err := dir.File.Readdir(0) // Readdir(0)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		if fi.IsDir() {
			err = cr.processDir(gofilepath.JointSmart(cr.dstPathSeparator, srcFilePath, fi.Name()), fi)
			if err != nil {
				if cr.IgnErr {
					log.Warnf("processDir error [local ignore]: %v", err)
				} else {
					return err
				}
			}
		} else {
			err = cr.copyFile(gofilepath.JointSmart(cr.dstPathSeparator, srcFilePath, fi.Name()), fi)
			if err != nil {
				if cr.IgnErr {
					log.Warnf("sendFile error [local ignore]: %v", err)
				} else {
					return err
				}
			}
		}
	}
	return err
}

// ParseTemplate parses a text template from a reader and applies it to the data and writes the output to the writer.
func ParseTemplate(r io.Reader, data interface{}, w io.Writer) (err error) {
	var b []byte
	if b, err = io.ReadAll(r); err != nil {
		return
	}

	t, err := template.New("template").Parse(string(b))
	if err != nil {
		return err
	}
	t = t.Option("missingkey=default")
	return t.Execute(w, data)
}

func (cr *CopyRecursive) copyFile(srcPath string, srcFileInfo os.FileInfo, mods ...fs.FileMode) (err error) {
	var relpath string
	relpath, err = gofilepath.Rel(cr.srcPath, srcPath)
	if err != nil {
		return
	}
	var r *File
	r, err = cr.Open(srcPath)
	if err != nil {
		return
	}
	defer r.Close()
	dstPath := gofilepath.JointSmart(cr.dstPathSeparator, cr.dstPath, relpath)
	fmode := srcFileInfo.Mode() | 0200
	if len(mods) != 0 {
		fmode = mods[0]
	}
	if cr.allFileMode != 0 {
		fmode = cr.allFileMode
	}
	// var envs map[string]string
	re, _ := regexp.Compile(cr.matchFilNameRegexpForTemplate)
	if cr.isTemplateMode && (cr.matchFilNameRegexpForTemplate == "" || (re != nil && re.MatchString(srcPath))) {
		var buff bytes.Buffer
		var dataI interface{}
		if cr.dataTmpl != nil {
			dataI = cr.dataTmpl
		} else {
			dataI = cr.envs
		}
		if err = ParseTemplate(r, dataI, &buff); err == nil {
			// err = t.Execute(cr.Writer, envs)
			err = cr.Writer(dstPath, &buff, fmode)
		}
	} else {
		err = cr.Writer(dstPath, r, fmode)
	}
	if err != nil {
		return
	}

	return err
}

func (cr *CopyRecursive) MkdirAll(dstPath string, dstFileInfo fs.FileMode) (err error) {
	eles := strings.Split(dstPath, cr.dstPathSeparator)
	paths := ""
	for i := 0; i < len(eles); i++ {
		if len(paths) == 0 {
			paths = eles[i]
		} else if eles[i] != "" {
			paths = strings.Join([]string{paths, eles[i]}, cr.dstPathSeparator)
		} else {
			continue
		}
		err = cr.mkdir(paths, dstFileInfo)
		if err != nil {
			break
		}
	}
	return
}

// fs.FileMode bool
// Copy srcName to dstName, srcName and dstName can be file or directory,
// mods is used to modify file mode, it can be one of the following values:
//
// 1. fs.FileMode or os.FileMode: this value will be applied to all files and
// directories being copied.
//
// 2. Slice of fs.FileMode or os.FileMode: each element in the slice will be
// applied to the file or directory that is being copied, in the order of the
// slice. For example, if the slice is []fs.FileMode{0644, 0755}, the first
// element (0644) will be applied to the first file or directory being copied,
// and the second element (0755) will be applied to the second file or
// directory being copied, and so on.
// the files in the directory, if it is a single value, it will be applied
// to the file or directory that is being copied.
//
// Note that if mods is a slice, the order of the elements in the slice is
// important, the first element in the slice will be applied to the file or
// directory that is being copied first, and so on.
func (cr *CopyRecursive) Copy(dstName, srcName string, mods ...interface{}) (err error) {
	if dstName == "" {
		return errors.New("dstName cannot empty")
	}

	var srcFileInfo fs.FileInfo
	srcFileInfo, err = cr.Stat(srcName)
	if err != nil {
		return err
	}
	cr.dstPathSeparator = gofilepath.GetPathSeparator(dstName)
	if cr.dstPathSeparator == "" {
		cr.dstPathSeparator = string(os.PathSeparator)
	}

	cr.srcPathSeparator = gofilepath.GetPathSeparator(srcName)
	var fmode fs.FileMode
	var hasMod bool
	cr.envs = os.Environ()
	cr.envs = []string{}
	for _, i := range mods {
		switch v := i.(type) {
		case fs.FileMode:
			fmode = v
			cr.allFileMode = v
			hasMod = true
		case uint32:
			fmode = fs.FileMode(v)
			cr.allFileMode = fmode
			hasMod = true
		case int:
			fmode = fs.FileMode(v)
			cr.allFileMode = fmode
			hasMod = true
		case map[string]string:
			cr.envs = make([]string, 0)
			for _, v := range v {
				if parts := strings.SplitN(v, "=", 2); len(parts) == 2 {
					cr.envs = append(cr.envs, fmt.Sprintf("%s=%s", parts[0], parts[1]))
				}
			}
		case bool:
			cr.isTemplateMode = v
		case string:
			cr.matchFilNameRegexpForTemplate = v
		default:
			cr.dataTmpl = v
		}
	}
	if srcFileInfo.IsDir() {
		if gofilepath.HasEndPathSeparators(dstName) {
			err = cr.mkdir(dstName, srcFileInfo.Mode()|0200)
			if err != nil {
				return
			}
		}

		// if !gofilepath.HasEndPathSeparators(srcName) {
		// if !gofilepath.HasEndPathSeparators(dstName) {
		// 	srcName += cr.srcPathSeparator
		// 	dstName += cr.dstPathSeparator
		// 	srcFileInfo, _ = cr.Stat(srcName)
		// }
		// 	dstName = gofilepath.JointSmart(cr.dstPathSeparator, dstName, gofilepath.Base(srcName))
		// 	err = cr.mkdir(dstName, srcFileInfo.Mode()|0200)
		// 	if err != nil {
		// 		return
		// 	}
		// }
	}
	cr.srcPath = srcName
	cr.dstPath = dstName

	if srcFileInfo.IsDir() {
		cr.srcPath = srcName
		err = cr.processDir(srcName, srcFileInfo)
		if err != nil {
			if cr.IgnErr {
				log.Warnf("error [ignore]: %v", err)
			} else {
				return
			}
		}
	} else {
		if hasMod {
			err = cr.copyFile(srcName, srcFileInfo, fmode)
		} else {
			err = cr.copyFile(srcName, srcFileInfo)

		}
		return err
	}
	return
}
