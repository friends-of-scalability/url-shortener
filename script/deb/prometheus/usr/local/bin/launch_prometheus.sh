#!/bin/bash
#This is simple bash script that is used to test access to the EC2 Parameter store.

function get_dependencies() {
    # Install the AWS CLI
    apt-get -y install python2.7 curl jq
    wget -qO- https://get.docker.com/ | sh
    curl -o /tmp/get-pip.py https://bootstrap.pypa.io/get-pip.py
    python2.7 /tmp/get-pip.py
    pip install awscli docker-compose
    # Getting region
}

function set_prometheus_config() {
    STUDENT=$(curl -s http://169.254.169.254/latest/meta-data/iam/info | jq .InstanceProfileArn | egrep -o 'student-\w+' | cut -f2 -d'-')
    sed -e "s/%student_id%/${STUDENT}/g" prometheus.yml.template > prometheus.yml
}


get_dependencies

cd /opt/prometheus/
set_prometheus_config
docker-compose up -d
