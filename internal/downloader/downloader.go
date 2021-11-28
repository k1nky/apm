package downloader

import (
	"io/ioutil"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/k1nky/apm/internal/move"
)

const (
	DefDestinationMode = 0755
)

type DownloaderGlobs []string

type DownloaderOptions struct {
	TempDirectory string
	Override      bool
	UseGitConfig  bool
	Globs         DownloaderGlobs
}

type Downloader struct {
	repo    *git.Repository
	options *DownloaderOptions
}

func (options *DownloaderOptions) Validate() (err error) {
	var tempDir string

	if options.TempDirectory == "" {
		if tempDir, err = ioutil.TempDir("", "apm-"); err != nil {
			return
		}
		options.TempDirectory = tempDir
	}
	if _, err = os.Stat(options.TempDirectory); err != nil {
		return
	}
	if len(options.Globs) == 0 {
		options.Globs = append(options.Globs, "*")
	}
	return nil
}

func (d *Downloader) clone(dir string, url string, ref *plumbing.Reference) (err error) {

	var (
		wt   *git.Worktree
		hash *plumbing.Hash
	)

	cloneOptions := &git.CloneOptions{
		URL:          url,
		SingleBranch: false,
		Tags:         git.AllTags,
	}
	if ref.Name().IsBranch() {
		cloneOptions.ReferenceName = ref.Name()
	}
	if d.repo, err = git.PlainClone(dir, false, cloneOptions); err != nil {
		return
	}

	if hash, err = d.repo.ResolveRevision(plumbing.Revision(ref.Name())); err != nil {
		return
	}
	if wt, err = d.repo.Worktree(); err != nil {
		return
	}
	err = wt.Checkout(&git.CheckoutOptions{
		Hash: *hash,
	})

	return
}

func (d *Downloader) pulse(url string, version string) (*plumbing.Reference, error) {
	remRepo := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
	})
	refs, err := remRepo.List(&git.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, ref := range refs {
		if ref.Name().Short() == version {
			return ref, nil
		}
	}

	return plumbing.NewHashReference(plumbing.ReferenceName(version), plumbing.NewHash(version)), nil
}

func (d *Downloader) Get(url string, version string, dest string, options *DownloaderOptions) error {

	if err := options.Validate(); err != nil {
		return err
	}
	d.options = options

	defer os.RemoveAll(options.TempDirectory)

	ref, err := d.pulse(url, version)
	if err != nil {
		return err
	}

	if err := d.clone(options.TempDirectory, url, ref); err != nil {
		return err
	}

	if err := move.Move(options.TempDirectory, dest, &move.MoveOptions{
		Globs:    options.Globs,
		Exclude:  []string{`\.git`},
		Override: options.Override,
	}); err != nil {
		return err
	}

	return nil
}
