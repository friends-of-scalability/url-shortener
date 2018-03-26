#!/usr/bin/env bash
set -x
MINIKUBE_BIN_DIRECTORY="$PWD/script/bin"
MINIKUBE_PATH="${MINIKUBE_BIN_DIRECTORY}/minikube"

function get_minikube() {
    mkdir -p ${MINIKUBE_BIN_DIRECTORY}
    $MINIKUBE_PATH version || curl -Lo $MINIKUBE_PATH https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && chmod +x $MINIKUBE_PATH
}

function check_or_start_cluster() {
    $MINIKUBE_PATH version
    $MINIKUBE_PATH status
    if [ $? -eq 0 ];then
        echo "minikube is already up and running"
    elif [ $? -eq 1 ];then
        echo "minikube is not running, starting up the cluster"
        $MINIKUBE_PATH start --kubernetes-version v1.8.0 --memory 4096 --cpus 2
        $MINIKUBE_PATH addons enable ingress
        $MINIKUBE_PATH addons enable registry
        $MINIKUBE_PATH addons enable heapster
    elif [ $? -eq 4 ];then
        echo "something went wrong in the cluster, check up logs"
    elif [ $? -eq 7 ];then
        echo "something went wrong in the cluster, check up logs"
    fi
}

get_minikube
check_or_start_cluster
