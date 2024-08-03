package export

import "strings"

type Exporter interface {
	Export() error
}

//go:generate stringer -type=Format -linecomment
type Format int

const (
	Invalid Format = iota // INVALID
	Text                  // txt
	JSON                  // json
	TOML                  // toml
	Table                 // table
	CSV                   // csv
	YAML                  // yml
)

func GetFormat(s string) Format {
	switch strings.ToLower(s) {
	case "json":
		return JSON
	case "toml":
		return TOML
	case "table":
		return Table
	case "csv":
		return CSV
	case "yaml", "yml":
		return YAML
	case "text", "txt":
		return Text
	default:
		return Invalid
	}
}
