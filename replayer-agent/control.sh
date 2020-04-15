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

# set web server
local_ip="127.0.0.1"
local_port="8998"

# TODO: tmp for private git
#git_go='https://github.com/didichuxing/sharingan-go/raw/recorder'
git_go='https://git.xiaojukeji.com/nuwa/tools/go/raw/master'

# build
function build() {
    if [ -z $GOROOT  ];then
        # install golang1.10
        curl "$git_go/install/go1.10" | sh
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at installing golang1.10, please check!!!\n"
            exit 1
        fi
        export GOROOT=/tmp/recorder-go1.10
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at setting GOROOT, please check!!!\n"
            exit 1
        fi
        export PATH=$GOROOT/bin:$PATH
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at updating PATH, please check!!!\n"
            exit 1
        fi
    fi

    which glide &> /dev/null
    if [ $? -ne 0 ]; then
        # install glide
        curl https://glide.sh/get | sh
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at installing glide, please check!!!\n"
            exit 1
        fi
    fi

    if [ -z $GOPATH  ];then
        prePath="/src/github.com/didichuxing/sharingan"
        if [[ $workspace == *$prePath* ]];then
            export GOPATH=`echo ${workspace%/src/*}`
            if [ $? -ne 0 ]; then
                printf "${error_msg}build failed at setting GOPATH, please check!!!\n"
                exit 1
            fi
        else
            printf "${error_msg}build failed at setting GOPATH, please check and clone sharingan to legal path!!!\n"
            exit 1
        fi
    fi

    cd $root
    rm -rf vendor
    glide install
    if [ $? -ne 0 ]; then
        printf "${error_msg}build failed at executing glide install, please check!!!\n"
        exit 1
    fi

    cd $workspace
    go build -o $app main.go
    if [ $? -ne 0 ]; then
        printf "${error_msg}build failed at go build $app, please check!!!\n"
        exit 1
    fi

    printf "${info_msg}$app builds success!!!\n"
}

# start
function start() {
    if [ ! -f "$agent_file" ]; then
        printf "${error_msg}please exesute sh control.sh build first!!!\n"
        exit 1
    fi

    if [[ "$OSTYPE" =~ ^linux.* ]]; then
        local_ip=`ip addr | grep "inet " | grep -v 127 | awk '{print $2}' | awk -F '/' '{print $1}'`
    fi

    ps -ef | grep $app | grep -v grep &> /dev/null
    if [ $? -eq 0 ];then
        printf "${warn_msg} $app has already started!!! Have fun with http://${local_ip}:${local_port}\n"
        exit 0
    fi

    # default -parallel=10
    nohup $agent_file -cursor >> $agent_log_file 2>&1 &
    sleep 2
    pid=$!

    ps -ef | grep $app | grep -v grep &> /dev/null
    if [ $? -ne 0 ];then
        printf "${error_msg} Failed to start $app, please check $agent_log_file and try again!!!\n"
        exit 1
    fi
    echo $pid > $agent_pid_file
    printf "${info_msg}$app starts success!!! [pid=${pid}]  Have fun with http://${local_ip}:${local_port} !\n"
}

# stop
function stop() {
    ps -ef | grep $app | grep -v grep &> /dev/null
    if [ $? -ne 0 ];then
        printf "${warn_msg}$app is not running!!!\n"
        exit 0
    fi

    pkill $app
    sleep 2

    ps -ef | grep $app | grep -v grep &> /dev/null
    if [ $? -eq 0 ];then
        printf "${error_msg}Failed to stop $app, please check and try again!!!\n"
        exit 1
    fi
    printf "${info_msg}$app stops success!!!\n"
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
