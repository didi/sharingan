#!/bin/bash

workspace=$(cd $(dirname $0) && pwd -P)
cd ${workspace}

app=replayer-agent
root=$(dirname $workspace)

# set file
agent_file=$workspace/$app
agent_pid_file=$workspace/agent.pid
agent_log_file=$workspace/agent.log

# set msg
info_msg="\033[;32m[INFO]\033[0m\t"
warn_msg="\033[;33m[WARN]\033[0m\t"
error_msg="\033[;31m[ERROR]\033[0m\t"

# build
function build() {
    if [ ! -d $GOROOT  ];then
        # install golang1.10
        curl https://github.com/didichuxing/sharingan-go/raw/recorder/install/go1.10 | sh
        if [ $? -ne 0 ]; then
            printf "${error_msg}build fail at installing golang1.10, please check!!!\n"
            exit 1
        fi
        export GOROOT=/tmp/recorder-go1.10
        if [ $? -ne 0 ]; then
            printf "${error_msg}build fail at setting GOROOT, please check!!!\n"
            exit 1
        fi
        export PATH=$GOROOT/bin:$PATH
        if [ $? -ne 0 ]; then
            printf "${error_msg}build fail at updating PATH, please check!!!\n"
            exit 1
        fi
    fi

    which glide > /dev/null 2>&1
    if [ $? -ne 0 ]; then
        # install glide
        curl https://glide.sh/get | sh
        if [ $? -ne 0 ]; then
            printf "${error_msg}build fail at installing glide, please check!!!\n"
            exit 1
        fi
    fi

    cd $root
    glide install
    if [ $? -ne 0 ]; then
        printf "${error_msg}build fail at executing glide install, please check!!!\n"
        exit 1
    fi

    cd $workspace
    go build -o $app main.go
    if [ $? -ne 0 ]; then
        printf "${error_msg}build fail at go build replayer-agent, please check!!!\n"
        exit 1
    fi

    printf "${info_msg}replayer-agent builds success!!!\n"
}

# start
function start() {
    if [ ! -f "$agent_file" ]; then
        printf "${error_msg}please exesute sh control.sh build first!!!\n"
        exit 1
    fi

    ps -ef | grep $app | grep -v grep &> /dev/null
    if [ $? -eq 0 ];then
        printf "${warn_msg} replayer-agent has already started!!!\n"
        exit 0
    fi

    # 默认值，如有需要可以修改
    export REPLAYER_MOCK_IP="127.0.0.1"
	  export REPLAYER_MOCK_PORT="3515"

    # default -parallel=10
    nohup $agent_file -cursor >> $agent_log_file 2>&1 &
    sleep 2
    pid=$!

    ps -ef | grep $app | grep -v grep &> /dev/null
    if [ $? -ne 0 ];then
        printf "${error_msg} Failed to start replayer-agent, please check and try again!!!\n"
        exit 1
    fi
    echo $pid > $agent_pid_file
    printf "${info_msg}replayer-agent starts success!!! [pid=${pid}]  Have fun with http://127.0.0.1:8998 !\n"
}

# stop
function stop() {
    ps -ef | grep $app | grep -v grep &> /dev/null
    if [ $? -ne 0 ];then
        printf "${warn_msg}replayer-agent is not running!!!\n"
        exit 0
    fi

    pkill $app
    sleep 2

    ps -ef | grep $app | grep -v grep &> /dev/null
    if [ $? -eq 0 ];then
        printf "${error_msg}Failed to stop replayer-agent, please check and try again!!!\n"
        exit 1
    fi
    printf "${info_msg}replayer-agent stops success!!!\n"
}

case $1 in
    "build" )
        build
        ;;
    "start" )
        start
        ;;
    "stop" )
        stop
        ;;
    "reload" )
        stop
        start
        ;;
    * )
        printf "${warn_msg}unknow command!!! [usage: build、start、stop、reload]\n"
        exit 1
        ;;
esac
