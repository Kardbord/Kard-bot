#!/bin/bash

DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/incl.sh"

initialize

MUTEX_PATH="mutex"
go tool pprof "$(getServer)${MUTEX_PATH}"
echo "done"
