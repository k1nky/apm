package move

import (
	"log"
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
)

var testSrcDir string

func setUp() (tmpdir string, err error) {
	tmpdir, err = os.MkdirTemp("", "apm-move-src")
	if err != nil {
		return
	}
	if _, err = git.PlainClone(tmpdir, false, &git.CloneOptions{
		URL: "https://bitbucket.org/bitjackass/public-roles",
	}); err != nil {
		return
	}
	return tmpdir, nil
}

func TestValidate(t *testing.T) {
	opt := new(MoveOptions)
	if err := opt.Validate(); err != nil {
		t.Error(err)
		return
	}
	if len(opt.Globs) == 0 {
		t.Error("expected at least wildcard glob")
	}
}

func TestMove(t *testing.T) {

}

func TestPickup(t *testing.T) {
	cases := map[string]struct {
		want int
		opt  *MoveOptions
	}{
		"all": {4, &MoveOptions{
			Globs:   []string{"*"},
			Exclude: []string{`\.git`},
		}},
		"*.yml": {2, &MoveOptions{
			Globs: []string{"*.yml"},
		}},
		"tasks": {1, &MoveOptions{
			Globs: []string{"tasks"},
		}},
		"doublestar": {2, &MoveOptions{
			Globs: []string{"**/*.yml"},
		}},
	}
	for k, v := range cases {
		t.Log(k)
		files, err := pickup(testSrcDir, v.opt)
		t.Log(files)
		if err != nil {
			t.Error(err)
			return
		}
		if len(files) != v.want {
			t.Errorf("expected %d matched files", v.want)
		}
	}
}

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
