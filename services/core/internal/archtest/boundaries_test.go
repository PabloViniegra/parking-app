package archtest

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

func TestInternalPackagesRespectBoundaries(t *testing.T) {
	root := ".."

	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return err
		}
		rel, err := filepath.Rel(root, filepath.Dir(p))
		if err != nil {
			return err
		}
		src, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		violations, err := CheckSource(path.Join(strings.TrimSuffix(internalPrefix, "/"), filepath.ToSlash(rel)), p, src)
		if err != nil {
			return err
		}
		for _, v := range violations {
			t.Error(v)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal packages: %v", err)
	}
}
