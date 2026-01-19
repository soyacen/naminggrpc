#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

echo "--- go generate start ---"
go generate ./...
echo "--- go generate end ---"

