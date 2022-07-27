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
	// Mode0755 is file mode 0755
	Mode0755 = 0755
	Mode0644 = 0644
)

type GlobOptions struct {
	// Exclude describes list of files which will be excluded from glob resolution
	Exclude []string
}

type CopyOptions struct {
	// Override existed destination directory
	Override bool
	// MakeLostDirectory creates parent directory if it is not exist
	MakeLostDirectory bool
}

func validateRoot(root string) (string, error) {
	if root == "" {
		return os.Getwd()
	}
	return root, nil
}

// ResolveGlob returns list of files within `root` and matched to `glob`
func ResolveGlob(root string, glob string, options *GlobOptions) (files []string, err error) {

	root, _ = validateRoot(root)
	if options == nil {
		options = &GlobOptions{}
	}
	if err := options.Validate(); err != nil {
		return nil, err
	}

	fs, err := filepath.Glob(path.Join(root, glob))
	if err != nil {
		return files, err
	}
	dirtyFiles := append(files, fs...)

	if len(options.Exclude) == 0 {
		files = dirtyFiles
	} else {
		for _, exclude := range options.Exclude {
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

// Validate copy options
func (options *CopyOptions) Validate() error {
	return nil
}

// Validate glob options. Directory .git is skipped by default.
func (options *GlobOptions) Validate() error {
	if len(options.Exclude) == 0 {
		options.Exclude = []string{".git"}
	}
	return nil
}

// CopyFile makes copy of `src` to `dest` and returns a number of copied bytes.
// `src` must be a regular file. Parent directory of `dest` must be existed.
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

// CopyDir makes recursive copy of a directory into another directory.
// If the destination directory does not exist it will be created.
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

// Copy makes copy of a file or directory (`src`) within a directory (`dest`).
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
		if options.MakeLostDirectory {
			err = os.MkdirAll(path.Dir(dest), Mode0755)
			if err != nil {
				return
			}
		}
		_, err = CopyFile(src, dest)
	}

	return
}
