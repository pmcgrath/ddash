#!/bin/bash
# Tried to do this with a make file but the scripting was just too painfull
set -e
#set -x

# See http://stackoverflow.com/questions/5143795/how-can-i-check-in-a-bash-script-if-my-local-git-repo-has-changes
if ! git diff-index --quiet HEAD --; then echo "YOU HAVE PENDING CHANGES !"; fi

declare -r commit_hash=`git rev-parse HEAD 2>/dev/null`
declare -r short_commit_hash=`git rev-parse --short HEAD 2>/dev/null`
declare -r build_date=`date +%FT%T%z`
declare -r ld_flags="-X main.commitHash ${commit_hash} -X main.shortCommitHash ${short_commit_hash} -X main.buildDate ${build_date}"

function clean() 
{
	go clean
}

function build()
{
	go build -ldflags "${ld_flags}"
}

function build-image()
{
	local sudo_run_prefix=
	if [ -z "$(groups | grep 'docker')" ]; then
		sudo docker build -t ${USER}/ddash:latest .
		sudo_run_prefix='sudo '
	else
		docker build -t ${USER}/ddash:latest .
	fi

	echo
	echo Use the following to run an instance within a container
	echo -e "\t${sudo_run_prefix}docker run --name=ddash --detach=true --volume=/var/run/docker.sock:/var/run/docker.sock:ro --publish=8090:8090 ${USER}/ddash:latest"
}

if [ $# == 0 ]; 		then build; build-image; fi
if [ "$1" == "build" ]; 	then build; fi
if [ "$1" == "buildimage" ]; 	then build-image; fi
if [ "$1" == "clean" ]; 	then clean; fi

