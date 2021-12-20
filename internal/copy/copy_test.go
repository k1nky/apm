package copy

import (
	"fmt"
	"log"
	"os"
	"path"
	"testing"

	"github.com/go-git/go-git/v5"
)

var testSrcDir string

func setUp() (tmpdir string, err error) {
	tmpdir, err = os.MkdirTemp("", "apm-copy-src")
	if err != nil {
		return
	}
	if _, err = git.PlainClone(tmpdir, false, &git.CloneOptions{
		URL:          "https://bitbucket.org/bitjackass/apm-test-example",
		SingleBranch: true,
	}); err != nil {
		return
	}
	return tmpdir, nil
}

func TestValidate(t *testing.T) {
	opt := new(CopyOptions)
	if err := opt.Validate(); err != nil {
		t.Error(err)
		return
	}
}

func testCopy(what string, want []string) (err error) {
	tmpdir, err := os.MkdirTemp("", "apm-copy-dest")
	if err != nil {
		return
	}
	defer os.RemoveAll(tmpdir)

	fs, err := ResolveGlob(testSrcDir, what, nil)
	if err != nil {
		return err
	}
	if err = Copy(testSrcDir, fs, tmpdir, nil); err != nil {
		return
	}
	for _, v := range want {
		if _, err := os.Stat(path.Join(tmpdir, v)); os.IsNotExist(err) {
			return fmt.Errorf("expected file %s does not exist", v)
		}
	}

	return
}

func TestCopyAll(t *testing.T) {
	want := []string{"sub_b/sub_b.yml", "sub_a/sub_a1/2.json",
		"sub_a/sub_a1/1.json", "sub_a/sub_a.yml", "main.txt"}
	if err := testCopy("*", want); err != nil {
		t.Error(err)
	}
}

func TestCopySubdir(t *testing.T) {
	want := []string{"sub_a/sub_a1/2.json", "sub_a/sub_a1/1.json"}
	if err := testCopy("sub_a/sub_a1", want); err != nil {
		t.Error(err)
	}
}

func TestCopyWildcard(t *testing.T) {
	want := []string{"sub_a/sub_a1/2.json", "sub_a/sub_a1/1.json"}
	if err := testCopy("sub_a/sub_a1/*.json", want); err != nil {
		t.Error(err)
	}
}

// TODO: Seems doublestart is not supported
// func TestCopyDoblestart(t *testing.T) {
// 	want := []string{"sub_a/sub_a1/2.json", "sub_a/sub_a1/1.json"}
// 	if err := testCopy("**/*.json", want); err != nil {
// 		t.Error(err)
// 	}
// }

func TestMain(m *testing.M) {
	var err error
	testSrcDir, err = setUp()
	if err != nil {
		log.Fatal(err)
		return
	}
	m.Run()
	os.RemoveAll(testSrcDir)
}
