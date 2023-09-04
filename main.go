// Package main is the entry point for the CLI
package main

import (
	"github.com/asphaltbuffet/elf/cmd"
)

var (
	version = "dev"
	date    = "unknown"
)

func main() {
	cmd.Execute(version, date)
}
