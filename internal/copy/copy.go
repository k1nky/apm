package copy

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

const (
	Mode0755 = 0755
)

type GlobOptions struct {
	Exclude []string
}

type CopyOptions struct {
	Override bool
}

func (opt *CopyOptions) Validate() error {
	return nil
}

func (opts *GlobOptions) Validate() error {
	if len(opts.Exclude) == 0 {
		opts.Exclude = []string{".git"}
	}
	return nil
}

func ValidateRoot(root string) (string, error) {
	if root == "" {
		return os.Getwd()
	}
	return root, nil
}

func ResolveGlob(root string, glob string, opts *GlobOptions) (files []string, err error) {

	root, _ = ValidateRoot(root)
	if opts == nil {
		opts = &GlobOptions{}
	}
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	fs, err := filepath.Glob(path.Join(root, glob))
	if err != nil {
		return files, err
	}
	dirtyFiles := append(files, fs...)

	if len(opts.Exclude) == 0 {
		files = dirtyFiles
	} else {
		for _, exclude := range opts.Exclude {
			re, err := regexp.Compile(exclude)
			if err != nil {
				return files, err
			}
			for _, f := range dirtyFiles {
				if matched := re.MatchString(f); !matched {
					files = append(files, f)
				}
			}
		}
	}

	return
}

func CopyFile(src string, dest string) (int64, error) {
	info, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !info.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return 0, err
	}
	defer destFile.Close()
	nBytes, err := io.Copy(destFile, srcFile)

	return nBytes, err
}

func CopyDir(src string, dest string) (err error) {
	var fds []os.FileInfo
	var srcInfo os.FileInfo

	if srcInfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dest, srcInfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dest, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if _, err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func Copy(src string, dest string, options *CopyOptions) (err error) {
	var info fs.FileInfo

	if options == nil {
		options = &CopyOptions{}
	}
	if err = options.Validate(); err != nil {
		return
	}
	if _, err = os.Stat(dest); err == nil && options.Override {
		if err = os.RemoveAll(dest); err != nil {
			return err
		}
	}

	if info, err = os.Stat(src); err != nil {
		return
	}
	if info.Mode().IsDir() {
		err = CopyDir(src, dest)
	} else {
		_, err = CopyFile(src, dest)
	}

	return
}
