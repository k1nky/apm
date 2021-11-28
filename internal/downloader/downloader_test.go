package downloader

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const (
	DefTestPublicURL = "https://bitbucket.org/bitjackass/public-roles"
	DefAPMGetTmpDir  = "/tmp/apm-get"
)

func isTestFileExist(repo *git.Repository, filename string) (err error) {
	wt, err := repo.Worktree()
	if err != nil {
		return
	}
	_, err = wt.Filesystem.Stat(filename)
	return
}

func prepareCloneTest() (tmpdir string, d *Downloader) {
	tmpdir, err := ioutil.TempDir("", "apm-test-")
	if err != nil {
		log.Panic(err)
		return "", nil
	}
	return tmpdir, &Downloader{}
}

func testClone(what string, want string) (err error) {
	tmpdir, d := prepareCloneTest()
	if d == nil {
		return errors.New("a tmp directory was not created")
	}
	defer tearDown(tmpdir)

	ref, err := d.pulse(DefTestPublicURL, what)
	if err != nil {
		return err
	}
	if err = d.clone(tmpdir, DefTestPublicURL, ref); err != nil {
		return err
	}

	err = isTestFileExist(d.repo, want)

	return

}

func TestMain(m *testing.M) {
	m.Run()
}

func TestDownloaderOptions(t *testing.T) {
	opt := &DownloaderOptions{}
	if err := opt.Validate(); err != nil {
		t.Error(err)
		return
	}

	if _, err := os.Stat(opt.TempDirectory); opt.TempDirectory == "" || err != nil {
		t.Error("invalid temp directory", err)
	}
	defer os.RemoveAll(opt.TempDirectory)
	if len(opt.Globs) == 0 {
		t.Error("expected at least one glob")
	} else if opt.Globs[0] != "*" {
		t.Error("expected default wildcard glob")
	}
}

func TestDownloaderOptionsInvalidDirectory(t *testing.T) {
	opt := &DownloaderOptions{
		TempDirectory: "/tmp/apm-invalid-path",
	}
	if err := opt.Validate(); err == nil {
		t.Error("expected not nil error")
	}
}

func TestInvalidDownloaderOptions(t *testing.T) {
	opt := DownloaderOptions{}
	if err := opt.Validate(); err != nil {
		t.Error(err)
		return
	}

	if _, err := os.Stat(opt.TempDirectory); opt.TempDirectory == "" || err != nil {
		t.Error("invalid temp directory", err)
	}
	defer os.RemoveAll(opt.TempDirectory)
}

func TestCloneByTag(t *testing.T) {
	if err := testClone("v2.0", "is_v2.yml"); err != nil {
		t.Error(err)
	}
}

func TestCloneByHash(t *testing.T) {
	if err := testClone("20a5b29", "is_hash.yml"); err != nil {
		t.Error(err)
	}
}

func TestCloneByBranch(t *testing.T) {
	if err := testClone("dev", "is_dev.yml"); err != nil {
		t.Error(err)
	}
}

func TestCloneByHashBranch(t *testing.T) {
	if err := testClone("db72642", "is_dev.yml"); err != nil {
		t.Error(err)
	}
}

func TestPulseUnreachable(t *testing.T) {
	d := &Downloader{}
	ref, err := d.pulse(DefTestPublicURL+"-invalid", "v2.0")
	if ref != nil || err == nil {
		t.Errorf("expected nil reference and error but got %v %s", ref, err)
	}
}

func TestPulseByTag(t *testing.T) {
	d := &Downloader{}
	ref, err := d.pulse(DefTestPublicURL, "v2.0")
	t.Logf("Got: %v", ref)
	if err != nil {
		t.Error(err)
	}
	if ref == nil || (ref != nil && !ref.Name().IsTag()) {
		t.Error("expected tag type")
	}
}

func TestPulseByBranch(t *testing.T) {
	d := &Downloader{}
	ref, err := d.pulse(DefTestPublicURL, "dev")
	t.Logf("Got: %v", ref)
	if err != nil {
		t.Error(err)
	}
	if ref == nil || (ref != nil && !ref.Name().IsBranch()) {
		t.Error("expected branch type")
	}
}

func TestPulseByHash(t *testing.T) {
	d := &Downloader{}
	ref, err := d.pulse(DefTestPublicURL, "ec30b20")
	t.Logf("Got: %v", ref)
	if err != nil {
		t.Error(err)
	}
	if ref == nil || (ref != nil && ref.Type() != plumbing.HashReference) {
		t.Errorf("expected nil reference")
	}
}

func TestGet(t *testing.T) {
	d := &Downloader{}
	cases := map[string]struct {
		version string
		opt     *DownloaderOptions
		wants   []string
	}{
		"default":     {"2.0", &DownloaderOptions{}, []string{"defaults", "tasks", "is_v2.yml"}},
		"tasks":       {"2.0", &DownloaderOptions{Globs: []string{"tasks"}}, []string{"tasks"}},
		"tasks/*.yml": {"2.0", &DownloaderOptions{Globs: []string{"tasks/*.yml"}}, []string{"tasks", "tasks/main.yml"}},
	}
	for k, v := range cases {
		t.Logf("test case %s", k)
		if err := d.Get(DefTestPublicURL, v.version, DefAPMGetTmpDir, v.opt); err != nil {
			t.Error(err)
			t.Fail()
		}
		files, err := filepath.Glob(path.Join(DefAPMGetTmpDir, "*"))
		if err != nil {
			t.Fail()
			t.Error(err)
		}
		sort.Strings(files)
		t.Log("got files ", files)
		for _, f := range v.wants {
			if idx := sort.SearchStrings(files, path.Join(DefAPMGetTmpDir, f)); idx == len(files) {
				t.Errorf("expected file %s not found", f)
				t.Fail()
			}
		}
		os.RemoveAll(DefAPMGetTmpDir)
	}
}

func tearDown(tmpdir string) {
	os.RemoveAll(tmpdir)
}
