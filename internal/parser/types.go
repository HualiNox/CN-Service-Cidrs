package parser

type SourceType string

const (
	ClashList  SourceType = "clash-list"
	SourceCIDR SourceType = "cidr"
)

type Source struct {
	Type  SourceType `yaml:"type" validate:"required,oneof=clash-list cidr"`
	Value string     `yaml:"value" validate:"required"`
}

type SourceGroup struct {
	Name    string   `yaml:"name" validate:"required"`
	Sources []Source `yaml:"sources" validate:"required"`
}

type SourceFile struct {
	Directory string
	Group     SourceGroup
}
