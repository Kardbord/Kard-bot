#!/bin/bash
# shellcheck disable=SC2155

REPO="kardbot"

function usage() {
  echo "${0}"
  echo
  echo "SYNOPSIS"
  echo "${0} {-M|--major|-m|--minor|-p|--patch} {-h|--help}"
  echo
  echo "DESCRIPTION"
  echo "Creates a ${REPO} release tarball containing a docker-compose file and any local mounts it requires."
  echo "Semantic versioning is handled automatically by the script based on the type of revision specified."
  echo
  echo "OPTIONS"
  echo "-M,--major  Indicates release is a major (potentially non-backwards compatible) release"
  echo "-m,--minor  Indicates release is a minor (backwards compatibile with new features) release"
  echo "-p,--patch  Indicates release is a patch (backwards compatible bugfix) release"
  echo "-t,--tag    Optionally provide a tag (e.g., alpha or beta) to append to the semantic version."
  echo "--push      Optionally push the created release tag to GitHub. If you select this option, you"
  echo "            should create an official release in the GitHub repo using the tarball produced by"
  echo "            this script. Otherwise that would be pretty naughty. :( Creating an official release"
  echo "            will trigger a GitHub action to build and publish an accompanying docker image."
  echo
  echo "EXAMPLES"
  echo "${0} -M"
  echo "${0} --major"
  echo "${0} --major -t alpha"
}

RED='\033[0;31m'
NC='\033[0m' # No Color
function fail() {
  if [ -n "${1}" ]; then
    echo -e "${RED}${1}${NC}"
  fi
  exit 1
}

if ! OPTS=$(getopt -n "${0}" -o Mmphpt: -l major,minor,patch,help,push,tag: -- "$@"); then
  usage
  exit 1
fi
eval set -- "$OPTS"

ISMAJOR=0
ISMINOR=0
ISPATCH=0
SHOULDPUSH=0
USERTAG=""
while true; do
  case "${1}" in
    -M|--major)
      ISMAJOR=1
      shift
      ;;
    -m|--minor)
      ISMINOR=1
      shift
      ;;
    -p|--patch)
      ISPATCH=1
      shift
      ;;
    --push)
      SHOULDPUSH=1
      shift
      ;;
    -t|--tag)
      USERTAG="-${2}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    --)
      break
      ;;
  esac
done

if [[ $((ISMAJOR+ISMINOR+ISPATCH)) -ne 1 ]]; then
  fail "must specify exactly one of --major, --minor, or --patch"
fi

if [[ $(git branch --show-current) != "main" ]]; then
  fail "Cannot create a release while on branch \"$(git branch --show-current)\". Can only release from branch \"main\"."
fi

function prompt_continue() {
  local prompt="Are you sure you want to continue?"
  if [ -n "${1}" ]; then
    prompt="${1}"
  fi
  read -p "${prompt}" -n 1 -r
  echo    # (optional) move to a new line
  [[ $REPLY =~ ^[Yy]$ ]] || exit 0
}

function branch_status() {
  if ! git diff-index --cached --quiet HEAD; then
    echo uncommitted-changes
  else
    local a="main" b="origin/main"
    local base=$( git merge-base $a $b )
    local aref=$( git rev-parse  $a )
    local bref=$( git rev-parse  $b )
    if [[ $aref == "$bref" ]]; then
      echo up-to-date
    elif [[ $aref == "$base" ]]; then
      echo behind
    elif [[ $bref == "$base" ]]; then
      echo ahead
    else
      echo diverged
    fi
  fi
}
git fetch > /dev/null
status=$(branch_status)
if [[ "${status}" != "up-to-date" ]]; then
  prompt_continue "Local branch differs from the remote with status: ${status}. Do you want to continue?"
fi

function build_release() {
  local tag="${1}"
  if [ "${tag}" == "" ]; then
    fail "No tag specified, there is a problem with this script."
  fi

  prompt_continue "Ready to build release ${tag}! Proceed?"
  git tag "${tag}" -m "${REPO} release ${tag}" || fail "Failed to create git tag ${tag}"

  local releasefile="${REPO}-${tag}.tar"
  echo "Tarring files..."
  tar -cvf "./${releasefile}" --xform "s|^./|${REPO}-${tag}/|" "./docker-compose.yml" "./config/" "./pasta/" || fail "Failed to create release ${releasefile}"
  tar -rvf "./${releasefile}" --xform "s|^./.env_example|${REPO}-${tag}/.env|" "./.env_example" || fail "Failed to append .env_example to archive"
  echo "Compressing files..."
  gzip -9 "./${releasefile}"
  if [ "${SHOULDPUSH}" -eq 1 ]; then
    git push origin "${tag}" || fail "Failed to push tag \"${tag}\"to GitHub"
    echo "Successfully created release ${releasefile} and pushed ${tag} to GitHub."
    echo "You should create an official release on the GitHub using ${releasefile}."
  else
    echo "Successfully created release ${releasefile}"
  fi
}

latest_tag=$(git describe --tags --match "v[0-9]*" --abbrev=0 2> /dev/null)

if [ "${latest_tag}" == "" ]; then
  # no tags found, this is the first release
  build_release "v0.0.0${USERTAG}"
  exit 0
fi

[[ "${latest_tag}" =~ ^v[0-9]+.[0-9]+.[0-9]+.* ]] || fail "Latest found tag (${latest_tag}) does not appear to be a SemVer, there is a problem with this script."

function parse_semver() {
  local token="$1"
  local major=0
  local minor=0
  local patch=0

  if grep -E '^v[0-9]+\.[0-9]+\.[0-9]+' <<<"$token" >/dev/null 2>&1 ; then
    # It has the correct syntax.
    local n="${token:1}"
    n="${n//[!0-9]/ }"
    local a=("${n//\./ }")
    major=${a[0]}
    minor=${a[1]}
    patch=${a[2]}
  fi

  echo "$major $minor $patch"
}

mapfile -d ' ' -t v < <(parse_semver "${latest_tag}")
MAJOR=${v[0]}
MINOR=${v[1]}
PATCH=${v[2]}
[[ $((MAJOR+MINOR+PATCH)) -ne 0 ]] || fail "Latest found tag (${latest_tag}) does not appear to be a SemVer, there is a problem with this script."

if [ ${ISMAJOR} -eq 1 ]; then
  MAJOR=$((MAJOR+1))
  MINOR=0
  PATCH=0
elif [ ${ISMINOR} -eq 1 ]; then
  MINOR=$((MINOR+1))
  PATCH=0
elif [ ${ISPATCH} -eq 1 ]; then
  PATCH=$((PATCH+1))
fi

fulltag="v${MAJOR}.${MINOR}.${PATCH}${USERTAG}"
if ! grep -q "${fulltag}" docker-compose.yml; then
  fail "Looks like docker-compose.yml hasn't been updated in preparation for version ${fulltag} yet."
fi

build_release "${fulltag}"
