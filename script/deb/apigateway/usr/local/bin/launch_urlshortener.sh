#!/bin/bash
#This is simple bash script that is used to test access to the EC2 Parameter store.
function unset_if_null_or_empty() {
    if [[ -z "${!1// }" || ""${!1}"" == "null" ]];then
        unset "${1}"
    fi
}

function get_parameter_store(){
    local VAR=$1
    local REGION=$2
    echo $(aws ssm get-parameters --names $VAR  --with-decryption --region $REGION --output json | jq .Parameters[0].Value | tr -d '"')
}

function get_environmental_variables() {
    EC2_AVAIL_ZONE=$(curl -s http://169.254.169.254/latest/meta-data/placement/availability-zone)
    EC2_REGION=$(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region)
    STUDENT=$(curl -s http://169.254.169.254/latest/meta-data/iam/info | jq .InstanceProfileArn | egrep -o 'student-\w+' | cut -f2 -d'-')

    # Trying to retrieve parameters from the EC2 Parameter Store
    export URLSHORTENER_POSTGRESQL_HOST=$(get_parameter_store "/"$STUDENT"/prod/db/host" $EC2_REGION)
    export URLSHORTENER_POSTGRESQL_USER=$(get_parameter_store "/"$STUDENT"/prod/db/user" $EC2_REGION)
    export URLSHORTENER_POSTGRESQL_PASSWORD=$(get_parameter_store "/"$STUDENT"/prod/db/password" $EC2_REGION)
    export URLSHORTENER_POSTGRESQL_PORT=$(get_parameter_store "/"$STUDENT"/prod/db/port" $EC2_REGION)
    export URLSHORTENER_FAKELOAD=$(get_parameter_store "/"$STUDENT"/prod/fakeload " $EC2_REGION)
    export URLSHORTENER_ROLE=$(cat /etc/urlshortener/role)
    export URLSHORTENER_STORAGE=$(get_parameter_store "/"$STUDENT"/prod/storage" $EC2_REGION)
    export URLSHORTENER_HTTP_ADDR=$(get_parameter_store "/"$STUDENT"/prod/http/addr" $EC2_REGION)
    export URLSHORTENER_SD_RESOLVER=$(get_parameter_store "/"$STUDENT"/prod/sd/resolver" $EC2_REGION)
    export URLSHORTENER_SD_SHORTENER=$(get_parameter_store "/"$STUDENT"/prod/sd/shortener" $EC2_REGION)
    export URLSHORTENER_ZIPKIN_ADDRESS=$(get_parameter_store "/"$STUDENT"/prod/zipkin/address" $EC2_REGION)
}

get_environmental_variables
unset_if_null_or_empty "URLSHORTENER_HTTP_ADDR"
unset_if_null_or_empty "URLSHORTENER_STORAGE"
unset_if_null_or_empty "URLSHORTENER_FAKELOAD"
unset_if_null_or_empty "URLSHORTENER_POSRGRESQL_HOST"
unset_if_null_or_empty "URLSHORTENER_POSTGRESQL_PORT"
unset_if_null_or_empty "URLSHORTENER_POSTGRESQL_USER"
unset_if_null_or_empty "URLSHORTENER_POSTGRESQL_PASSWORD"
unset_if_null_or_empty "URLSHORTENER_ROLE"
unset_if_null_or_empty "URLSHORTENER_SD_RESOLVER"
unset_if_null_or_empty "URLSHORTENER_SD_SHORTENER"

/opt/url-shortener/bin/url-shortener
