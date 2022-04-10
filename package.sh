#!/bin/bash
# shellcheck disable=SC2155

pushd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null || fail "Failed to enter script directory"

REPO="kardbot"

function usage() {
  echo "${0}"
  echo
  echo "SYNOPSIS"
  echo "  ${0} {-h|--help}"
  echo
  echo "DESCRIPTION"
  echo "  Creates a ${REPO} release tarball containing a docker-compose file and any local mounts it requires."
  echo
  echo "OPTIONS"
  echo "  -h, --help Prints this help message."
}

RED='\033[0;31m'
NC='\033[0m' # No Color
function fail() {
  if [ -n "${1}" ]; then
    echo -e "${RED}${1}${NC}"
  fi
  exit 1
}

if ! OPTS=$(getopt -n "${0}" -o h -l help -- "$@"); then
  usage
  exit 1
fi
eval set -- "$OPTS"

while true; do
  case "${1}" in
  -h | --help)
    usage
    exit 0
    ;;
  --)
    break
    ;;
  esac
done

function build_release() {
  local tag="${1}"
  if [ "${tag}" == "" ]; then
    fail "No tag specified, there is a problem with this script."
  fi

  local releasefile="${REPO}-${tag}.tar"
  local tmpcompose="$(mktemp)"
  trap 'rm -f $tmpcompose' EXIT
  cat docker-compose.yml > "${tmpcompose}"
  sed -i "s|kardbot:latest|kardbot:${tag}|g" "${tmpcompose}"
  sed -i "s|kardbot-latest|kardbot-${tag}|g" "${tmpcompose}"
  echo "Tarring files..."
  tar -cvf "./${releasefile}" --xform "s|^./|${REPO}-${tag}/|" "./LICENSE" "./README.md" "./config/" "./assets/" &>/dev/null || fail "Failed to create release ${releasefile}"
  tar -rvf "./${releasefile}" --xform "s|^./.env_example|${REPO}-${tag}/.env|" "./.env_example" &>/dev/null || fail "Failed to append .env_example to archive"
  tar -rvf "${releasefile}" --xform "s|^.*$|${REPO}-${tag}/docker-compose.yml|" "${tmpcompose}" &>/dev/null || fail "Failed to append docker-compose.yml to archive"
  echo "Compressing files..."
  gzip -9 "./${releasefile}"
  echo "Successfully created release ${releasefile}"
}

build_release "$(git describe --tags --abbrev=0 2>/dev/null)"
