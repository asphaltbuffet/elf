#!/bin/bash
set -euo pipefail

rm -rf manpages
mkdir manpages

go run . man
gzip -9 manpages/*.1
