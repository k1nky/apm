package parser

type SrcDest struct {
	Src  string `yaml:"src"`
	Dest string `yaml:"dest"`
}

type Boost struct {
	Copy []SrcDest `yaml:"copy"`
}

type Apkg struct {
	Mappings []SrcDest `yaml:"mappings"`
	Boost    Boost     `yaml:"boost"`
}

type Mapping struct {
	SrcDest
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
