package parser

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Mapping struct {
	Src     string `yaml:"src"`
	Dest    string `yaml:"dest"`
	Version string `yaml:"version"`
	Boost   bool   `yaml:"boost"`
}

type Package struct {
	Src      string    `yaml:"src"`
	Mappings []Mapping `yaml:"mappings"`
}

type Requirements struct {
	Packages []Package `yaml:"packages"`
}

func Load(r io.Reader) (*Requirements, error) {
	var (
		cfg   *Requirements
		err   error
		bytes []byte
	)

	if bytes, err = ioutil.ReadAll(r); err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &Requirements{}
	}

	// TODO: validate requirements

	return cfg, err
}
