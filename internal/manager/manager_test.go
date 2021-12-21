package manager

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setUp() (tmpdir string, err error) {
	tmpdir, err = os.MkdirTemp("", "apm-project")
	return
}

func testInstallPackage(p *Package) (err error) {
	m := Manager{}
	if m.WorkDir, err = setUp(); err != nil {
		return err
	}
	defer os.RemoveAll(m.WorkDir)
	if err = m.Install(p, &InstallOptions{}); err != nil {
		return
	}
	err = filepath.Walk(m.WorkDir, func(path string, info fs.FileInfo, err error) error {
		if !strings.Contains(path, ".apm") {
			if info.Mode()&os.ModeSymlink != 0 {
				if _, err := filepath.EvalSymlinks(path); err != nil {
					return err
				}
			}
		}
		return nil
	})

	return
}

func TestInstallDefault(t *testing.T) {
	p := &Package{
		URL:      "https://bitbucket.org/bitjackass/apm-test-example",
		Version:  "master",
		Path:     ".",
		Mappings: []Mapping{{"*", "."}},
	}
	if err := testInstallPackage(p); err != nil {
		t.Error(err)
	}
}

func TestInstallSubdir(t *testing.T) {
	p := &Package{
		URL:      "https://bitbucket.org/bitjackass/apm-test-example",
		Version:  "master",
		Path:     ".",
		Mappings: []Mapping{{"sub_a", "project/a"}},
	}
	if err := testInstallPackage(p); err != nil {
		t.Error(err)
	}
}

func TestInstallDir(t *testing.T) {
	p := &Package{
		URL:      "https://bitbucket.org/bitjackass/apm-test-example",
		Version:  "master",
		Path:     ".",
		Mappings: []Mapping{{".", "project"}},
	}
	if err := testInstallPackage(p); err != nil {
		t.Error(err)
	}
}

func TestInstallSubfiles(t *testing.T) {
	p := &Package{
		URL:      "https://bitbucket.org/bitjackass/apm-test-example",
		Version:  "master",
		Path:     ".",
		Mappings: []Mapping{{"sub_a/sub_a1/*.json", "project"}},
	}
	if err := testInstallPackage(p); err != nil {
		t.Error(err)
	}
}
