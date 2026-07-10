package runners

import _ "embed"

// GoTemplate is the built-in Go runner wrapper template.
//
//go:embed interface/go.tmpl
var GoTemplate []byte

// PythonTemplate is the built-in Python runner wrapper template.
//
//go:embed interface/python.tmpl
var PythonTemplate []byte

// BashTemplate is the built-in Bash runner wrapper template.
//
//go:embed interface/bash.tmpl
var BashTemplate []byte

// RustTemplate is the built-in Rust runner wrapper template (rendered to src/main.rs).
//
//go:embed interface/rust.tmpl
var RustTemplate []byte

// CSharpTemplate is the built-in C# runner wrapper template (rendered to runtime-wrapper.cs).
//
//go:embed interface/csharp.tmpl
var CSharpTemplate []byte

// F77Template is the built-in Fortran 77 runner wrapper template (a C harness compiled with gfortran).
//
//go:embed interface/f77.tmpl
var F77Template []byte

// LuaTemplate is the built-in Lua runner wrapper template.
//
//go:embed interface/lua.tmpl
var LuaTemplate []byte
