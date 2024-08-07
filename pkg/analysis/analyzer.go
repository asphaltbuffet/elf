package analysis

import "strings"

type Analyzer interface {
	Graph(GraphType) error
	Stats() error
}

//go:generate stringer -type=GraphType
type GraphType int

const (
	Invalid GraphType = iota
	Line
	Box
)

func StringToGraphType(s string) GraphType {
	switch strings.ToLower(s) {
	case "line":
		return Line
	case "box":
		return Box
	default:
		return Invalid
	}
}
