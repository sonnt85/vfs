package vfs

import (
	"archive/zip"
	"embed"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/afero/zipfs"

	"github.com/sonnt85/vfs/sembed"
)

// import
func NewEmbedHttpSystemFS(efs *embed.FS, rootDir string, sub ...string) (*VFS, error) {
	return NewEFs(efs, rootDir, sub...)
}

// rootDir need manual name efs
func NewEFs(efs *embed.FS, rootDir string, sub ...string) (*VFS, error) {
	ef := &sembed.EFs{
		FS: efs,
	}

	if len(rootDir) == 0 {
		return nil, fmt.Errorf("rootDir can not empty")
	}
	if len(sub) != 0 {
		rootDir = filepath.Join(rootDir, sub[0])
	}
	return NewVFS(ef, rootDir)
}

func NewVFSFromAFS(afs afero.Fs) *VFS {
	return &VFS{
		VFSS: &VFSS{
			afs,
		},
	}
}

// base, layer are afero.Fs or *vfs.VFS
func NewOverlayFs(base, layer interface{}) *VFS {
	var baseafs, layerafs afero.Fs
	// var efi interface{}
	switch v := base.(type) {
	case afero.Fs:
		baseafs = v
	case *VFS:
		baseafs = v.VFI
	default:
		return nil
	}

	switch v := layer.(type) {
	case afero.Fs:
		layerafs = v
	case *VFS:
		layerafs = v.VFI
	default:
		return nil
	}
	return NewVFSFromAFS(afero.NewCopyOnWriteFs(baseafs, layerafs))
	// if baseafs, ok := base.(afero.Fs); ok {
	// 	if layerafs, ok := layer.(afero.Fs); ok {
	// 		return NewVFSFromAFS(afero.NewCopyOnWriteFs(baseafs, layerafs))
	// 	}
	// }
	// return nil
}

// base, layer are afero.Fs or *vfs.VFS
func NewOsFs(dir ...string) *VFS {
	af := afero.NewOsFs()
	if len(dir) != 0 {
		af = afero.NewBasePathFs(af, dir[0])
	}
	return NewVFSFromAFS(af)
}

// r is  *zip.Reader) or stringpath to zipfile
func NewZipFs(r interface{}) (vfs *VFS, err error) {
	var rd *zip.Reader
	zippath := ""
	switch v := r.(type) {
	case *zip.Reader:
		rd = v
	case string:
		var rdc *zip.ReadCloser
		rdc, err = zip.OpenReader(zippath)
		if err != nil {
			return
		}
		var ok bool
		if rd, ok = interface{}(rdc).(*zip.Reader); ok {
			err = fmt.Errorf("r not implement *zip.Reader")
			return
		}
	}
	return NewVFSFromAFS(zipfs.New(rd)), nil
}
