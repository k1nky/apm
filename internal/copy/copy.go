package copy

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

const (
	DefaultDestinationMode = 0755
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

func copyFiles(root string, files []string, dest string, options *CopyOptions) error {
	if err := os.MkdirAll(dest, DefaultDestinationMode); err != nil {
		return err
	}
	for _, f := range files {
		relpath, _ := filepath.Rel(root, f)
		destFile := path.Join(dest, relpath)
		_, err := os.Stat(destFile)
		if os.IsNotExist(err) {
			os.MkdirAll(path.Dir(destFile), DefaultDestinationMode)
			srcInfo, _ := os.Stat(f)
			if srcInfo.IsDir() {
				if err := copyDir(f, destFile); err != nil {
					return err
				}
			} else {
				if _, err := copyFile(f, destFile); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func copyFile(src string, dest string) (int64, error) {
	sfs, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sfs.Mode().IsRegular() {
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

func copyDir(src string, dest string) (err error) {
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
			if err = copyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if _, err = copyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func Copy(root string, files []string, dest string, options *CopyOptions) error {
	if options == nil {
		options = &CopyOptions{}
	}
	if err := options.Validate(); err != nil {
		return err
	}
	if _, err := os.Stat(dest); os.IsExist(err) && options.Override {
		if err := os.RemoveAll(dest); err != nil {
			return err
		}
	}

	if err := copyFiles(root, files, dest, options); err != nil {
		return err
	}

	return nil
}