#!/usr/bin/env bash

set -e

if [ "${PACKAGE_NAME_PREFIX}" == "" ]; then
    echo "PACKAGE_NAME_PREFIX environment variable value is not valid."
    exit 1
fi

NAME=${PACKAGE_NAME_PREFIX}-url-shortener-$1
DESCRIPTION="Friends of scalability $NAME"
WORKING_PATH="$(dirname ${0})"

# Set dummy version, if not set already (e.g. outside of CI)
if [[ ! ${VERSION} ]]; then
  VERSION=0.$(date +%Y%m%d%H%M%S)
  echo "WARN: Setting dummy VERSION to: ${VERSION}"
fi

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
PKG_BUILD_DIR="/tmp/rpm.${RANDOM}"; mkdir "${PKG_BUILD_DIR}"

mkdir -p ${PKG_BUILD_DIR}/opt/url-shortener/bin/

if [[ "$NAME" != "hystrixdashboard" || "$NAME" != "zipkin"  || "$NAME" != "prometheus" ]];then
  cp bin/urlshortener ${PKG_BUILD_DIR}/opt/url-shortener/bin/url-shortener
fi

rsync -av script/deb/$1/ ${PKG_BUILD_DIR}/

pushd ${WORKING_PATH}
fpm \
  -s dir \
  -t deb \
  -n ${NAME} \
  -v ${VERSION} \
  --iteration=$(git rev-parse --short HEAD) \
  --description "${DESCRIPTION}" \
  -d "stress" \
  -d "python2.7" \
  -d "curl" \
  -d "jq" \
  -d "python-pip" \
  --after-install ${PKG_BUILD_DIR}/usr/local/bin/after_install.sh \
  -C ${PKG_BUILD_DIR}
popd
