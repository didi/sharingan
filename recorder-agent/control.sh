#!/bin/bash

workspace=$(cd $(dirname $0) && pwd -P)
cd ${workspace}

# set file
agent_file=$workspace/agent
agent_pid_file=$workspace/agent.pid
agent_log_file=$workspace/agent.log

# set msg
info_msg="\033[;32m[INFO]\033[0m\t"
warn_msg="\033[;33m[WARN]\033[0m\t"
error_msg="\033[;31m[ERROR]\033[0m\t"

# build
function build() {
    if [ ! -d $GOROOT  ];then
        printf "${error_msg}build fail, please set GOROOT first!!!\n"
        exit 1  
    fi

    go build -o agent main.go
    if [ $? -ne 0 ]; then
        printf "${error_msg}build fail, please check!!!\n"
        exit 1
    fi

    printf "${info_msg}build success!!!\n"
}

# start
function start() {
    if [ ! -f "$agent_file" ]; then
        printf "${error_msg}please build first!!!\n"
        exit 1
    fi

    if [ -f "$agent_pid_file" ]; then
        printf "${warn_msg}already start!!!\n"
        exit 1
    fi

    nohup $agent_file >> $agent_log_file 2>&1 &
    pid=$!

    echo $pid > $agent_pid_file
    printf "${info_msg}start success!!! [pid=${pid}]\n"
}

# stop
function stop() {
    if [ ! -f "$agent_pid_file" ]; then 
        printf "${warn_msg}not start!!!\n"
        exit 1
    fi

    pid=`cat $agent_pid_file`
    if [ ! -z "$pid" ]; then
        kill $pid &>/dev/null
        sleep 1
        rm -rf $agent_pid_file
    fi
    printf "${info_msg}stop success!!!\n"
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