#!/bin/bash
set -euo pipefail

rm -rf completions
mkdir completions

for sh in bash zsh fish; do
	go run . completion "$sh" >"completions/elf.$sh"
done
