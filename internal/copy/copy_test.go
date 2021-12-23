package copy

import (
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

func TestCopyDir(t *testing.T) {
	what := []string{"sub_b", "."}
	want := [][]string{
		{"sub_b/sub_b.yml"},
		{"sub_b/sub_b.yml", "sub_a/sub_a1/2.json",
			"sub_a/sub_a1/1.json", "sub_a/sub_a.yml", "main.txt"},
	}

	for k, v := range what {
		t.Run(v, func(t *testing.T) {
			tmpdir, err := os.MkdirTemp("", "apm-copy-dest")
			if err != nil {
				return
			}
			defer os.RemoveAll(tmpdir)

			if err := Copy(path.Join(testSrcDir, v), path.Join(tmpdir, v), nil); err != nil {
				t.Error(err)
				return
			}
			for _, w := range want[k] {
				if _, err := os.Stat(path.Join(tmpdir, w)); os.IsNotExist(err) {
					t.Errorf("expected file %s does not exist", w)
				}
			}

		})
	}
}

func TestCopyFile(t *testing.T) {
	what := []string{"main.txt", "sub_a/sub_a.yml"}
	want := [][]string{
		{"main.txt"},
		{"sub_a/sub_a.yml"},
	}

	for k, v := range what {
		t.Run(v, func(t *testing.T) {
			tmpdir, err := os.MkdirTemp("", "apm-copy-dest")
			if err != nil {
				return
			}
			defer os.RemoveAll(tmpdir)

			dest := path.Join(tmpdir, v)
			os.MkdirAll(path.Dir(dest), Mode0755)
			if err := Copy(path.Join(testSrcDir, v), dest, nil); err != nil {
				t.Error(err)
				return
			}
			for _, w := range want[k] {
				if _, err := os.Stat(path.Join(tmpdir, w)); os.IsNotExist(err) {
					t.Errorf("expected file %s does not exist", w)
				}
			}

		})
	}
}

// TODO: TestResolveGlob
// want := []string{"sub_b/sub_b.yml", "sub_a/sub_a1/2.json",
// "sub_a/sub_a1/1.json", "sub_a/sub_a.yml", "main.txt"}

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
