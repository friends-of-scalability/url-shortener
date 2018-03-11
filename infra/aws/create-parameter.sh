#!/usr/bin/env bash

set -eu

if [ "${#}" -lt 3 ]
then
	echo "usage: ${0} <PARAMETER_NAME> <PARAMETER_TYPE> <PARAMETER_VALUE>" >&2
	echo >&2
	exit 1
fi

NAME=$1
TYPE=$2
VALUE=$3

aws ssm put-parameter --name ${NAME} --type ${TYPE} --value ${VALUE} --overwrite
