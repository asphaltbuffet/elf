//go:build tools

package tools // import "github.com/asphaltbuffet/elf/tools"

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint" // golangci-lint
	_ "github.com/vektra/mockery/v2"                        // mockery
	_ "golang.org/x/tools/cmd/stringer"                     // stringer
	_ "gotest.tools/gotestsum"                              // gotestsum
)
