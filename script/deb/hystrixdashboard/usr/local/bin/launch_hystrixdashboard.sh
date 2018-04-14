#!/bin/bash
#This is simple bash script that is used to test access to the EC2 Parameter store.

function get_dependencies() {
    wget -qO- https://get.docker.com/ | sh
    pip install docker-compose
    # Getting region
}


get_dependencies
cd /opt/hystrixdashboard/
docker-compose up -d
