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
	DefTestPublicURL = "https://bitbucket.org/bitjackass/apm-test-example"
	DefAPMGetTmpDir  = "/tmp/apm-get"
)

type testFileFingerprint struct {
	Hash string
	Path string
}

var (
	cloneCase map[string][]testFileFingerprint = map[string][]testFileFingerprint{
		"dev": {
			{"4f260daf425fbcb28450a73b4875b2c3", "sub_b/sub_b.yml"},
			{"ded11434ce59b89da43803f40599c25f", "sub_a/sub_a1/2.json"},
			{"798ec063a2559fe9182c13a1801a72a2", "sub_a/sub_a1/1.json"},
			{"29fc9c2a655b2713fc456f3caa5c7aa4", "sub_a/sub_a.yml"},
			{"e77989ed21758e78331b20e477fc5582", "main.txt"},
		},
		"v2.0": {
			{"4f260daf425fbcb28450a73b4875b2c3", "sub_b/sub_b.yml"},
			{"99914b932bd37a50b983c5e7c90ae93b", "sub_a/sub_a1/2.json"},
			{"798ec063a2559fe9182c13a1801a72a2", "sub_a/sub_a1/1.json"},
			{"29fc9c2a655b2713fc456f3caa5c7aa4", "sub_a/sub_a.yml"},
			{"8b4e9455dfd3b112a055967deff47ea2", "main.txt"},
		},
		"6954198": {
			{"8084392a5c5ea785d82d66efbdd118f4", "sub_b/sub_b.yml"},
			{"ded11434ce59b89da43803f40599c25f", "sub_a/sub_a1/2.json"},
			{"798ec063a2559fe9182c13a1801a72a2", "sub_a/sub_a1/1.json"},
			{"29fc9c2a655b2713fc456f3caa5c7aa4", "sub_a/sub_a.yml"},
			{"71ccb7a35a452ea8153b6d920f9f190e", "main.txt"},
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
	return tmpdir, NewDownloader()
}

func testGet(version string, want []testFileFingerprint) (err error) {
	tmpdir, d := prepareCloneTest()
	if d == nil {
		return errors.New("a tmp directory was not created")
	}
	defer tearDown(tmpdir)

	if _, err = d.Get(DefTestPublicURL, version, tmpdir, &Options{}); err != nil {
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
	if err := testGet("v2.0", cloneCase["v2.0"]); err != nil {
		t.Error(err)
	}
}

func TestGetByHash(t *testing.T) {
	if err := testGet("6954198", cloneCase["6954198"]); err != nil {
		t.Error(err)
	}
}

func TestGetByBranch(t *testing.T) {
	if err := testGet("dev", cloneCase["dev"]); err != nil {
		t.Error(err)
	}
}

func tearDown(tmpdir string) {
	os.RemoveAll(tmpdir)
}
