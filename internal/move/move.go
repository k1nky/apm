package move

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
)

const (
	DefDestinationMode = 0755
)

type MoveOptions struct {
	Globs    []string
	Exclude  []string
	Override bool
}

func (opt *MoveOptions) Validate() error {
	if len(opt.Globs) == 0 {
		opt.Globs = append(opt.Globs, "*")
	}
	return nil
}

func pickup(src string, options *MoveOptions) (files []string, err error) {
	dirtyFiles := []string{}
	for _, glob := range options.Globs {
		fs, err := filepath.Glob(path.Join(src, glob))
		if err != nil {
			return files, err
		}
		dirtyFiles = append(files, fs...)
	}
	if len(options.Exclude) == 0 {
		files = dirtyFiles
	}
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
	return
}

func move(src string, dest string, files []string, options *MoveOptions) error {
	if err := os.MkdirAll(dest, DefDestinationMode); err != nil {
		return err
	}
	for _, f := range files {
		original, _ := filepath.Rel(src, f)
		destFile := path.Join(dest, original)
		_, err := os.Stat(destFile)
		if os.IsNotExist(err) {
			os.MkdirAll(path.Dir(destFile), DefDestinationMode)
			os.Rename(f, destFile)
		}
	}

	return nil
}

func Move(src string, dest string, options *MoveOptions) error {
	if options == nil {
		options = new(MoveOptions)
	}
	if err := options.Validate(); err != nil {
		return err
	}
	if _, err := os.Stat(dest); os.IsExist(err) && options.Override {
		if err := os.RemoveAll(dest); err != nil {
			return err
		}
	}

	files, err := pickup(src, options)
	if err != nil {
		return err
	}
	if err := move(src, dest, files, options); err != nil {
		return err
	}

	return nil
}
