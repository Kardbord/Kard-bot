#!/bin/bash

DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/incl.sh"

initialize

BLOCK_PATH="block"
go tool pprof "$(getServer)${BLOCK_PATH}"
echo "done"
