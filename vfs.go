package vfs

import (
	"archive/tar"
	"archive/zip"
	"os"

	"embed"
	"fmt"
	"path/filepath"

	"github.com/sonnt85/gosutils/endec"
	"github.com/sonnt85/vfs/sembed"

	"github.com/spf13/afero"
	"github.com/spf13/afero/tarfs"
	"github.com/spf13/afero/zipfs"
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

func NewVFSFromAFS(afs afero.Fs, sub ...string) (vfs *VFS) {
	if len(sub) != 0 {
		afs = afero.NewBasePathFs(afs, sub[0])
	}
	vfs = new(VFS)
	vfs.AferoWrap = new(AferoWrap)
	vfs.AferoWrap.Afero = new(afero.Afero)
	vfs.AferoWrap.Afero.Fs = afs
	return
}

// base, layer are afero.Fs or *vfs.VFS
func NewOverlayFs(base, layer interface{}) *VFS {
	var baseafs, layerafs afero.Fs
	// var efi interface{}
	switch v := base.(type) {
	case afero.Fs:
		baseafs = v
	case *VFS:
		baseafs = v.Afero.Fs
	default:
		return nil
	}

	switch v := layer.(type) {
	case afero.Fs:
		layerafs = v
	case *VFS:
		layerafs = v.Afero.Fs
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
func NewZipFs(r interface{}, password ...[]byte) (vfs *VFS, err error) {
	var rd *zip.Reader
	switch v := r.(type) {
	case *zip.Reader:
		rd = v
	case string:
		var rdc *zip.ReadCloser

		rdc, err = zip.OpenReader(v)
		if err != nil {
			return
		}
		if len(password) != 0 {
			rdc.RegisterDecompressor(endec.ZIPTYPEAES, endec.NewAesDecrypter(password[0], endec.AESZIPCHUNKSIZE))
		}
		// fmt.Println("zipdirroot ", rdc.File[0].Name)
		// rd = &rdc.Reader
		// if len(rdc.File) == 2 && rdc.File[0].Mode().IsDir() && !rdc.File[1].FileInfo().IsDir() {
		// 	var tf fs.File
		// 	if tf, err = rdc.Open(rdc.File[1].Name); err == nil {
		// 		var trd *tar.Reader
		// 		trd = tar.NewReader(tf)
		// 		if _, err = trd.Next(); err == nil {
		// 			// fmt.Println("zipdirroot ", rdc.File[0].Name)
		// 			return NewVFSFromAFS(tarfs.New(trd)), nil
		// 		}
		// 	}
		// 	tf.Close()
		// }

		// rdc.Open(rdc.File[1].Name)
		// for _, v := range rdc.File {
		// 	fmt.Println(v.Name)
		// }
		return NewVFSFromAFS(zipfs.New(&rdc.Reader), rdc.File[0].Name), nil
	}
	return NewVFSFromAFS(zipfs.New(rd)), nil
}

// r is  *tar.Reader) or stringpath to tarfile
func NewTarFs(r interface{}, password ...[]byte) (vfs *VFS, err error) {
	var rd *tar.Reader
	switch v := r.(type) {
	case *tar.Reader:
		rd = v
	case string:
		var tf *os.File
		tf, err = os.Open(v)
		if err != nil {
			return
		}
		rd = tar.NewReader(tf)
		if _, err = rd.Next(); err != nil {
			tf.Close()
			return
		}
		// fmt.Println("zipdirroot ", rdc.File[0].Name)
		return NewVFSFromAFS(tarfs.New(rd)), nil
	}
	return NewVFSFromAFS(tarfs.New(rd)), nil
}

func NewArchiveFs(r interface{}, password ...[]byte) (vfs *VFS, err error) {
	if vfs, err = NewTarFs(r); err == nil {
		return
	}
	return NewZipFs(r, password...)
}
