#!/usr/bin/env bash
SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
$SCRIPT_DIR/docker-build.sh ci
mv ${SCRIPT_DIR}/*.deb ${SCRIPT_DIR}/..
deb-s3 upload --bucket geekshubs-cto-artifacts --prefix deb --arch amd64 --codename trusty --preserve-versions true *.deb
