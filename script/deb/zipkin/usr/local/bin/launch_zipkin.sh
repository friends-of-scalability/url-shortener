#!/bin/bash
#This is simple bash script that is used to test access to the EC2 Parameter store.

function get_dependencies() {
    # Install the AWS CLI
    apt-get -y install python2.7 curl jq python-pip
    wget -qO- https://get.docker.com/ | sh
    pip install awscli docker-compose
    # Getting region
}


get_dependencies
cd /opt/zipkin/
docker-compose up -d
