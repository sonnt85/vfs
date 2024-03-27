package vfs

import (
	"embed"
	"fmt"
	"testing"

	// "github.com/sonnt85/gosutils/sembed"

	"github.com/spf13/afero"
)

//go:embed statictest/**
var es embed.FS

func TestVFS(t *testing.T) {
	efs, _ := NewEFs(&es, "statictest")
	// fmt.Printf(efs)
	fmt.Println(efs.FindFilesMatchRegexpPathFromRoot("/", "file", 5, true, true))
	f := afero.NewOsFs()
	vf, _ := NewVFS(f, "/tmp")
	fmt.Println(vf.FindFilesMatchRegexpPathFromRoot("/", "kkk", 10, true, false))
	vf.Copy("/tmp/abcde", "tmpold/kkk", 0777)
	fmt.Println(vf.ReadFile("/tmpold/kkk"))
}
