package parser

type SourceType string

const (
	ClashList  SourceType = "clash-list"
	SourceCIDR SourceType = "cidr"
	CountryCSV SourceType = "country-csv"
	DomainList SourceType = "domain"
)

type Source struct {
	Type        SourceType `yaml:"type" validate:"required,oneof=clash-list cidr country-csv domain"`
	Value       string     `yaml:"value" validate:"required"`
	CountryCode string     `yaml:"country_code" validate:"required_if=Type country-csv,omitempty,len=2,uppercase"`
}

type SourceGroup struct {
	Name    string   `yaml:"name" validate:"required"`
	Sources []Source `yaml:"sources" validate:"required,min=1,dive"`
}

type SourceFile struct {
	Directory string
	Group     SourceGroup
}
