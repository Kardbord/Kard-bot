#!/bin/bash

DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/incl.sh"

initialize

HEAP_PATH="heap"
go tool pprof "$(getServer)${HEAP_PATH}"
echo "done"
