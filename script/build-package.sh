#!/usr/bin/env bash

set -e

NAME=url-shortener
DESCRIPTION="Friends of scalability url shortener"
WORKING_PATH="$(dirname ${0})"

# Set dummy version, if not set already (e.g. outside of CI)
if [[ ! ${VERSION} ]]; then
  VERSION=0.$(date +%Y%m%d%H%M%S)
  echo "WARN: Setting dummy VERSION to: ${VERSION}"
fi

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
PKG_BUILD_DIR="/tmp/rpm.${RANDOM}"; mkdir "${PKG_BUILD_DIR}"

mkdir -p ${PKG_BUILD_DIR}/opt/url-shortener/bin/

cp bin/urlshortener ${PKG_BUILD_DIR}/opt/url-shortener/bin/url-shortener

pushd ${WORKING_PATH}
fpm \
  -s dir \
  -t deb \
  -n ${NAME} \
  -v ${VERSION} \
  --iteration=$(git rev-parse --short HEAD) \
  --description "${DESCRIPTION}" \
  -C ${PKG_BUILD_DIR}
popd
