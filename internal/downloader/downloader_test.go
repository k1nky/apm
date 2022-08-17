package downloader

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"
)

const (
	DefTestPublicURL = "https://github.com/k1nky/ansible-simple-role.git"
	DefAPMGetTmpDir  = "/tmp/apm-get"
)

type testFileFingerprint struct {
	Hash string
	Path string
}

var (
	cloneCase map[string][]testFileFingerprint = map[string][]testFileFingerprint{
		"dev": {
			{"0be56bccb50804c305f4777f925b43bc", "defaults/main.yml"},
			{"b2bb3d0065e6282093a34942d73ff3e8", "tasks/main.yml"},
			{"97206668bdd762c00997fbcf308b5697", "templates/motd.j2"},
		},
		"v1.0": {
			{"0be56bccb50804c305f4777f925b43bc", "defaults/main.yml"},
			{"b2bb3d0065e6282093a34942d73ff3e8", "tasks/main.yml"},
			{"642fb044ab16285b7db45080fb0d5678", "templates/motd.j2"},
		},
		"39d3c976b8d06bb81f37e806ff915c05253d16ad": {
			{"0be56bccb50804c305f4777f925b43bc", "ansible-simple-role/defaults/main.yml"},
			{"b2bb3d0065e6282093a34942d73ff3e8", "ansible-simple-role/tasks/main.yml"},
			{"97206668bdd762c00997fbcf308b5697", "ansible-simple-role/templates/motd.j2"},
		},
	}
)

func getFileMD5(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}

	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func verifyDirectory(dir string, want []testFileFingerprint) (err error) {
	for _, v := range want {
		if h, err := getFileMD5(path.Join(dir, v.Path)); err != nil || h != v.Hash {
			return fmt.Errorf("file fingerprint is invalid or file does not exist")
		}
	}
	return nil
}

func prepareCloneTest() (tmpdir string, d *Downloader) {
	tmpdir, err := ioutil.TempDir("", "apm-test-")
	if err != nil {
		log.Panic(err)
		return "", nil
	}
	return tmpdir, NewDownloader(nil)
}

func testGet(version string, want []testFileFingerprint) (err error) {
	tmpdir, d := prepareCloneTest()
	if d == nil {
		return errors.New("a tmp directory was not created")
	}
	defer tearDown(tmpdir)

	if err = d.Get(DefTestPublicURL, version, tmpdir); err != nil {
		return
	}

	err = verifyDirectory(tmpdir, want)

	return
}

func TestMain(m *testing.M) {
	m.Run()
}

func TestDownloaderOptions(t *testing.T) {
	opt := &Options{}
	if err := opt.Validate(); err != nil {
		t.Error(err)
		return
	}
}

func TestGetByTag(t *testing.T) {
	if err := testGet("v1.0", cloneCase["v2.0"]); err != nil {
		t.Error(err)
	}
}

func TestGetByHash(t *testing.T) {
	if err := testGet("39d3c97", cloneCase["39d3c97"]); err != nil {
		t.Error(err)
	}
}

func TestGetByBranch(t *testing.T) {
	if err := testGet("dev", cloneCase["dev"]); err != nil {
		t.Error(err)
	}
}

func TestRewriteURL(t *testing.T) {
	what := []string{"repo.example.org", "http://repo.example.com", "ssh://192.168.1.1"}
	want := []string{"https://repo.example.org", "http://repo.example.com", "ssh://192.168.1.1"}
	for k, v := range what {
		if url, err := RewriteURL(v, false); url != want[k] || err != nil {
			t.Errorf("unexpected rewrite url from %s to %s", v, want[k])
		}
	}
}

func TestRewriteURLWithGitconfig(t *testing.T) {
	root, _ := os.Getwd()
	os.Setenv("XDG_CONFIG_HOME", path.Join(root, "./tests"))
	what := []string{"repo.example.org"}
	want := []string{"ssh://git@repo.example.org"}
	for k, v := range what {
		if url, err := RewriteURL(v, true); url != want[k] || err != nil {
			t.Errorf("unexpected rewrite url from %s to %s", v, want[k])
		}
	}
}

func TestFetchVersion(t *testing.T) {
	d := NewDownloader(nil)
	versions, err := d.FetchVersion("https://github.com/k1nky/ansible-simple-role.git")
	if err != nil {
		t.Error("unexpected error ", err)
	}
	if len(versions) == 0 {
		t.Error("there is no version")
	}
}

func tearDown(tmpdir string) {
	os.RemoveAll(tmpdir)
}
