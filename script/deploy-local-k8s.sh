#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
KUBE_LATEST_VERSION="v1.8.0"
HELM_VERSION="v2.8.2"
BIN_DIRECTORY="${SCRIPT_DIR}/bin"
HELM_PATH=${BIN_DIRECTORY}/helm
KUBECTL_PATH="${BIN_DIRECTORY}/kubectl"
CHARTS_PATH="${SCRIPT_DIR}/../infra/k8s/helm/charts"
RELEASE_NAME="dev"

function get_kubectl() {
    mkdir -p ${BIN_DIRECTORY}
    $KUBECTL_PATH version || curl -Lo $KUBECTL_PATH https://storage.googleapis.com/kubernetes-release/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl && chmod +x $KUBECTL_PATH
}

function get_helm() {
    mkdir -p ${BIN_DIRECTORY}
    if [[ ! -f /tmp/helm-${HELM_VERSION}.tar.gz ]];then
        curl --output /tmp/helm-${HELM_VERSION}.tar.gz https://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz
        mkdir /tmp/helm
        tar -zxvf /tmp/helm-${HELM_VERSION}.tar.gz -C /tmp/helm
    fi
    cp /tmp/helm/linux-amd64/helm ${HELM_PATH}
}

function package_and_deploy() {
    set +e
    ${KUBECTL_PATH} get deployment tiller-deploy --namespace kube-system &> /dev/null
    if [[ $? -ne 0 ]];then
        ${HELM_PATH} init --upgrade
        sleep 15
    fi
    set -e
    find ${CHARTS_PATH} -name "*.tgz" -delete
    ${HELM_PATH} package -u ${CHARTS_PATH}/url-shortener -d ${CHARTS_PATH}
    ${HELM_PATH} upgrade -i ${RELEASE_NAME} $(find ${CHARTS_PATH} -maxdepth 1 -name "*.tgz")
}

get_kubectl
get_helm
package_and_deploy
