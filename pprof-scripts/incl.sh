#!/bin/bash

# This script doesn't do anything on its own.
# It's for sharing code between the other scripts in this directory.

CFG="../config/setup.json"
PPROF_PATH="debug/pprof/"

function getServer {
  echo "http://$(pprofAddr)/${PPROF_PATH}"
}

function verifyDeps {
  if ! command -v jq &>/dev/null; then
    echo "You need to install jq: https://stedolan.github.io/jq/"
  fi
  if ! command -v curl &>/dev/null; then
    echo "You need to install curl: https://curl.se/"
  fi
  if ! command -v go &>/dev/null; then
    echo "You need to install go: https://go.dev/"
  fi
}

function pprofAddr {
  jq -r .pprof.address "${CFG}"
}

function verifyServerUp {
  curl -s -o /dev/null -w '%{http_code}' "$(getServer)"
}

function initialize {
  if [[ "$(verifyDeps)" != "" ]]; then
    verifyDeps
    exit 1
  fi

  echo "Checking for pprof server at $(getServer)"
  serverStatus=$(verifyServerUp)
  echo "pprof server status ${serverStatus}"
  if [[ "${serverStatus}" != "200" ]]; then
    echo "pprof server does not appear to be running"
    exit 1
  fi
}
