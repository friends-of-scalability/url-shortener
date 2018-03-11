#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
KUBE_LATEST_VERSION="v1.9.3"
HELM_VERSION="v2.8.2"
BIN_DIRECTORY="${SCRIPT_DIR}/bin"
HELM_PATH=${BIN_DIRECTORY}/helm
KUBECTL_PATH="${BIN_DIRECTORY}/kubectl"
CHARTS_PATH="${SCRIPT_DIR}/../helm/charts"

function get_kubectl() {
    mkdir -p ${BIN_DIRECTORY}
    $KUBECTL_PATH version || curl -Lo $KUBECTL_PATH https://storage.googleapis.com/kubernetes-release/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl && chmod +x $KUBECTL_PATH
}

function get_helm() {
    mkdir -p ${HELM_PATH}
    curl --output /tmp/helm.tar.gz https://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz
    tar -zxvf /tmp/helm.tar.gz -C /tmp \
    && mv /tmp/linux-amd64/helm ${HELM_PATH} \
    && rm -rf /tmp/linux-amd64

}

function package_and_deploy() {
    ${HELM_PATH}/helm init --upgrade
    sleep 15
    find ${CHARTS_PATH} -name "*.tgz" -delete
    ${HELM_PATH}/helm package ${CHARTS_PATH}/url-shortener -d ${CHARTS_PATH}
    ${HELM_PATH}/helm upgrade -i alumni-1 $(find ${CHARTS_PATH} -name "*.tgz")
}

get_kubectl
get_helm
package_and_deploy
