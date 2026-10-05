package parser

type SourceType string

const (
	ClashList SourceType = "clash-list"
)

type Source struct {
	Type  SourceType `yaml:"type" validate:"required,oneof=clash-list"`
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
