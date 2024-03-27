package vfs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"strings"

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

	IgnErr           bool
	srcPath          string
	dstPath          string
	srcPathSeparator string
	dstPathSeparator string
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
	err = cr.Writer(dstPath, r, fmode)
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

func (cr *CopyRecursive) Copy(dstName, srcName string, mods ...fs.FileMode) (err error) {
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
		err = cr.copyFile(srcName, srcFileInfo, mods...)
		return err
	}
	return
}
