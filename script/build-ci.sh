#!/usr/bin/env bash

sudo chown jenkins.jenkins * -R
rm *.deb

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
$SCRIPT_DIR/docker-build.sh ci
mv ${SCRIPT_DIR}/*.deb ${SCRIPT_DIR}/..

deb-s3 upload --s3-region eu-west-1 --bucket geekshubs-cto-artifacts --prefix deb --arch amd64 --codename trusty --preserve-versions true *.deb

