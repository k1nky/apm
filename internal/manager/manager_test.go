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
	tmpDir := ""
	if tmpDir, err = setUp(); err != nil {
		return err
	}
	defer os.RemoveAll(m.WorkDir)
	if err = m.Install([]*Package{p}, &InstallOptions{
		WorkDir: tmpDir,
	}); err != nil {
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
		URL:     "https://github.com/k1nky/ansible-simple-role.git",
		Version: "master",
		Src:     ".",
		Dest:    "roles/simple",
	}
	if err := testInstallPackage(p); err != nil {
		t.Error(err)
	}
}

func TestInstallSubdir(t *testing.T) {
	p := &Package{
		URL:     "https://github.com/k1nky/ansible-simple-roles.git",
		Version: "master",
		Src:     "motd",
		Dest:    "roles/motd",
	}
	if err := testInstallPackage(p); err != nil {
		t.Error(err)
	}
}

func TestInstallSubfiles(t *testing.T) {
	p := &Package{
		URL:     "https://github.com/k1nky/ansible-simple-role.git",
		Version: "master",
		Src:     "tasks/main.yml",
		Dest:    "project/main.yml",
	}
	if err := testInstallPackage(p); err != nil {
		t.Error(err)
	}
}
