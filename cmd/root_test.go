package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoot_HasEncryptDecrypt(t *testing.T) {
	root := GetRootCommand()
	var haveEncrypt, haveDecrypt bool
	for _, c := range root.Commands() {
		switch c.Name() {
		case "encrypt":
			haveEncrypt = true
		case "decrypt":
			haveDecrypt = true
		}
	}
	require.True(t, haveEncrypt, "encrypt command not registered")
	require.True(t, haveDecrypt, "decrypt command not registered")
}
