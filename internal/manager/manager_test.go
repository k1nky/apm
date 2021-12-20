package manager

import (
	"os"
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
	err = m.Install(p, &InstallOptions{})
	return
}

func TestInstallDefault(t *testing.T) {
	p := &Package{
		URL:     "https://bitbucket.org/bitjackass/apm-test-example",
		Version: "master",
		Path:    ".",
		Mappings: []struct {
			Src  string
			Dest string
		}{{"*", "."}},
	}
	if err := testInstallPackage(p); err != nil {
		t.Error(err)
	}
}

func TestInstallSubdir(t *testing.T) {
	p := &Package{
		URL:     "https://bitbucket.org/bitjackass/apm-test-example",
		Version: "master",
		Path:    ".",
		Mappings: []struct {
			Src  string
			Dest string
		}{{"sub_a", "project/a"}},
	}
	if err := testInstallPackage(p); err != nil {
		t.Error(err)
	}
}

func TestInstallDir(t *testing.T) {
	p := &Package{
		URL:     "https://bitbucket.org/bitjackass/apm-test-example",
		Version: "master",
		Path:    ".",
		Mappings: []struct {
			Src  string
			Dest string
		}{{".", "project"}},
	}
	if err := testInstallPackage(p); err != nil {
		t.Error(err)
	}
}

func TestInstallSubfiles(t *testing.T) {
	p := &Package{
		URL:     "https://bitbucket.org/bitjackass/apm-test-example",
		Version: "master",
		Path:    ".",
		Mappings: []struct {
			Src  string
			Dest string
		}{{"sub_a/sub_a1/*.json", "project"}},
	}
	if err := testInstallPackage(p); err != nil {
		t.Error(err)
	}
}
