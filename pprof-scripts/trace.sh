#!/bin/bash

DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/incl.sh"

initialize

PERIOD=30
TRACE_PATH="trace?seconds=${PERIOD}"
TRACE_FILE="trace.$(date +"%Y.%m.%d.%H.%M.%S").out"
wget -O "${TRACE_FILE}" "$(getServer)${TRACE_PATH}"
go tool trace "${TRACE_FILE}"
