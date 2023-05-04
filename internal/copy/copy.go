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
	Plain    bool
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

// CopyFile makes copy of `src` file to `dest` file and returns a number of copied bytes.
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
// If set `plain` to true content of `src` directory will be placed into `dest` directory.
func CopyDir(src string, dest string, plain bool) (err error) {
	var fds []os.FileInfo
	var srcInfo os.FileInfo

	if srcInfo, err = os.Stat(src); err != nil {
		return err
	}

	if !plain {
		dest = path.Join(dest, path.Base(src))
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
			if err = CopyDir(srcfp, dstfp, true); err != nil {
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
// Exmaples:
// Copy dir_a to dir_b and option Plain set to false => dir_a/dir_b
// Copy dir_a to dir_b and option Plain set to true => dir_b/<content of dir_a>
// Copy file_a to dir_b/file_b => dir_b/file_b
func Copy(src string, dest string, options *CopyOptions) (err error) {
	var info fs.FileInfo

	if options == nil {
		options = &CopyOptions{}
	}
	if err = options.Validate(); err != nil {
		return
	}

	if info, err = os.Stat(src); err != nil {
		return
	}
	if info.Mode().IsDir() {
		// Copy a directory
		if options.Override {
			if err = os.RemoveAll(dest); err != nil {
				return
			}
		}
		if err = os.MkdirAll(path.Dir(dest), Mode0755); err != nil {
			return
		}
		err = CopyDir(src, dest, options.Plain)
	} else {
		// Copy a file
		if options.Override {
			if err = os.RemoveAll(path.Dir(dest)); err != nil {
				return
			}
		}
		if err = os.MkdirAll(path.Dir(dest), Mode0755); err != nil {
			return
		}
		_, err = CopyFile(src, dest)
	}

	return
}
