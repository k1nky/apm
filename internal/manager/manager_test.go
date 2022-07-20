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

func testInstallFromApkg(p *Package) (err error) {
	m := Manager{}
	tmpDir := ""
	if tmpDir, err = setUp(); err != nil {
		return err
	}
	defer os.RemoveAll(m.WorkDir)
	if err = m.InstallFromApkg(p, &InstallOptions{
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
		URL:      "https://github.com/k1nky/ansible-simple-role.git",
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
		URL:      "https://github.com/k1nky/ansible-simple-roles.git",
		Version:  "master",
		Path:     ".",
		Mappings: []Mapping{{"motd", "roles/motd"}},
	}
	if err := testInstallPackage(p); err != nil {
		t.Error(err)
	}
}

func TestInstallDir(t *testing.T) {
	p := &Package{
		URL:      "https://github.com/k1nky/ansible-simple-roles.git",
		Version:  "master",
		Path:     ".",
		Mappings: []Mapping{{".", "roles"}},
	}
	if err := testInstallPackage(p); err != nil {
		t.Error(err)
	}
}

func TestInstallSubfiles(t *testing.T) {
	p := &Package{
		URL:      "https://github.com/k1nky/ansible-simple-role.git",
		Version:  "master",
		Path:     ".",
		Mappings: []Mapping{{"tasks/*.yml", "project"}},
	}
	if err := testInstallPackage(p); err != nil {
		t.Error(err)
	}
}

func TestInstallFromApkg(t *testing.T) {
	p := &Package{
		URL:     "https://github.com/k1nky/ansible-simple-roles.git",
		Version: "master",
		Path:    "motd",
	}
	if err := testInstallFromApkg(p); err != nil {
		t.Error(err)
	}
}
