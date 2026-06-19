package runners

import _ "embed"

// GoTemplate is the built-in Go runner wrapper template.
//
//go:embed interface/go.tmpl
var GoTemplate []byte

// PythonTemplate is the built-in Python runner wrapper template.
//
//go:embed interface/python.templ
var PythonTemplate []byte

// BashTemplate is the built-in Bash runner wrapper template.
//
//go:embed interface/bash.tmpl
var BashTemplate []byte
