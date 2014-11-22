#!/bin/bash

set -e
set -x

go build

sudo docker build -t pmcgrath/ddash:latest .
