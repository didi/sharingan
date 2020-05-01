#!/bin/bash

workspace=$(cd $(dirname $0) && pwd -P)
info_msg="\033[;32m[INFO]\033[0m\t"

echo "${info_msg}-> install go1.13"
curl -s https://raw.githubusercontent.com/didi/sharingan-go/recorder/install/go1.13 | sh > /dev/null 2>&1
export GOROOT=/tmp/recorder-go1.13
export PATH=$GOROOT/bin:$PATH

cd ${workspace}/example
echo "${info_msg}-> build example with recorder tag"
go build -tags="recorder"
if [ $? -ne 0 ]; then
    exit 1
fi

cd ${workspace}/example
echo "${info_msg}-> build example with replayer tag"
go build -tags="replayer" -gcflags="all=-N -l"
if [ $? -ne 0 ]; then
    exit 1
fi

cd ${workspace}/recorder-agent
echo "${info_msg}-> build recorder-agent"
go build
if [ $? -ne 0 ]; then
    exit 1
fi

cd ${workspace}/replayer-agent
echo "${info_msg}-> build replayer-agent"
go build
if [ $? -ne 0 ]; then
    exit 1
fi