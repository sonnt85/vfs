package vfs

import (
	"embed"
	"testing"

	"github.com/spf13/afero"
)

//go:embed statictest/**
var es embed.FS

func TestVFS(t *testing.T) {
	efs, err := NewEFs(&es, "statictest")
	if err != nil {
		t.Fatalf("NewEFs: %v", err)
	}
	files := efs.FindFilesMatchRegexpPathFromRoot("/", "file", 5, true, true)
	t.Logf("efs files: %v", files)

	vf, err := NewVFS(afero.NewOsFs(), "/tmp")
	if err != nil {
		t.Fatalf("NewVFS: %v", err)
	}
	found := vf.FindFilesMatchRegexpPathFromRoot("/", "kkk", 10, true, false)
	t.Logf("vf files: %v", found)

	// Best-effort smoke calls; missing paths should not fail the test.
	_ = vf.Copy("/tmp/abcde", "tmpold/kkk", 0777)
	_, _ = vf.ReadFile("/tmpold/kkk")
}
