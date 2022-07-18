package parser

type SrcDest struct {
	Src  string `yaml:"src"`
	Dest string `yaml:"dest"`
}

type ApkgBoost struct {
	Copy []SrcDest `yaml:"copy"`
}

type Apkg struct {
	Mappings []SrcDest `yaml:"mappings"`
	Boost    ApkgBoost `yaml:"boost"`
	Base     string    `yaml:"base"`
}

type ReqiuredMapping struct {
	SrcDest
	Version string `yaml:"version"`
}

type RequiredPackage struct {
	Url      string            `yaml:"src"`
	Mappings []ReqiuredMapping `yaml:"mappings"`
}

type Requirements struct {
	Packages []RequiredPackage `yaml:"packages"`
}
