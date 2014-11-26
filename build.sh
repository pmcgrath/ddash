#!/bin/bash
set -e
#set -x

go build

sudo_run_prefix=
if [ -z "$(groups | grep 'docker1')" ]; then
	sudo docker build -t ${USER}/ddash:latest .
	sudo_run_prefix='sudo '
else
	docker build -t ${USER}/ddash:latest .
fi

echo
echo Use the following to run an instance within a container
echo -e "\t${sudo_run_prefix}docker run --name=ddash --detach=true --volume=/var/run/docker.sock:/var/run/docker.sock:ro --publish=8090:8090 ${USER}/ddash:latest"

