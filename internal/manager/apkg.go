package manager

import (
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

type ApkgBoost struct {
	Copy []Mapping `yaml:"copy"`
}

type Apkg struct {
	Mappings []Mapping `yaml:"mappings"`
	Boost    ApkgBoost `yaml:"boost"`
}

// TODO: parse apkg
func ReadApkg(p string) (apkg *Apkg, err error) {

	// TODO: yaml == yml
	file, err := os.Open(path.Join(p, ".apkg.yml"))
	if os.IsNotExist(err) {
		return nil, fmt.Errorf(".apkg is not exist in the specified path")
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	apkg = &Apkg{}
	err = yaml.NewDecoder(file).Decode(apkg)

	// TODO: validateConfig(cfg)
	return
}
