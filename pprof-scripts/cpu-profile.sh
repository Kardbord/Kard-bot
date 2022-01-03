#!/bin/bash

DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/incl.sh"

initialize

PERIOD="30"
CPU_PATH="profile?seconds=${PERIOD}"
go tool pprof "$(getServer)${CPU_PATH}"
echo "done"
