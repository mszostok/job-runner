#
# Library of useful utilities for CI purposes.
#

readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly NC='\033[0m' # No Color

# Prints first argument as header. Additionally prints current date.
shout() {
    echo -e "
#################################################################################################
# $(date)
# $1
#################################################################################################
"
}

host::os() {
  local host_os
  case "$(uname -s)" in
    Darwin)
      host_os=osx
      ;;
    Linux)
      host_os=linux
      ;;
    *)
      echo "Unsupported host OS. Must be Linux or Mac OS X."
      exit 1
      ;;
  esac
  echo "${host_os}"
}

host::arch() {
  local host_arch
  case "$(uname -m)" in
    x86_64*)
      host_arch=x86_64
      ;;
    i?86_64*)
      host_arch=x86_64
      ;;
    amd64*)
      host_arch=x86_64
      ;;
    ppc64le*)
      host_arch=ppcle_64
      ;;
    *)
      echo "Unsupported host arch. Must be x86_64, or ppc64le."
      exit 1
      ;;
  esac
  echo "${host_arch}"
}
