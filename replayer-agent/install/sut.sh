#!/bin/bash

workspace=$(cd $(dirname $0) && pwd -P)
cd ${workspace}

app=example

# 需要制定
sharinganPath=~/devspace/gocommon/src/github.com/didichuxing/sharingan
installPath=$sharinganPath/replayer-agent/install

# set file
agent_file=$workspace/agent
agent_pid_file=$workspace/agent.pid
agent_log_file=$workspace/agent.log

# set msg
info_msg="\033[;32m[INFO]\033[0m\t"
warn_msg="\033[;33m[WARN]\033[0m\t"
error_msg="\033[;31m[ERROR]\033[0m\t"

# cov file
cov_time=`date "+%Y%m%d%H%M%S"`  
cov_start_file=/tmp/ShaRinGan/coverage.$app
cov_report_file=/tmp/ShaRinGan/coverage.$app.${cov_time}
cov_report_file_html=${cov_report_file}.html
cov_report_url="http://127.0.0.1:8998/coverage/report/coverage.$app.${cov_time}.html"

# build
function build() {
    # goroot
    if [ ! -d $GOROOT  ];then
        printf "${error_msg}build fail, please set GOROOT first!!!\n"
        exit 1  
    fi

    # build
    cp $installPath/codeCov/main_test.go $workspace/main_test.go
    go test \
        -gcflags="all=-N -l" -tags="replayer" \
        -v -c -covermode=count -coverpkg ./... \
        -o $workspace/$app.test
    
    # clear
    rm -rf $workspace/main_test.go
    
    printf "${info_msg}build success!!!\n"
}

# start
function start() {
    # check
    if [ ! -f "$workspace/$app.test" ]; then
        printf "${error_msg}please build first!!!\n"
        exit 1
    fi

    # check
    if [ -f "$agent_pid_file" ]; then
        printf "${warn_msg}already start!!!\n"
        exit 1
    fi

    # start
    nohup $workspace/$app.test -systemTest \
        -test.coverprofile=${cov_start_file} \
        > $agent_log_file 2>&1 &
    pid=$!

    echo $pid > $agent_pid_file
    printf "${info_msg}start success!!! [pid=${pid}]\n"
    printf "${info_msg}open http://127.0.0.1:8998 for covery test\n"
}

# stop
function stop() {
    # check
    if [ ! -f "$agent_pid_file" ]; then 
        printf "${warn_msg}not start!!!\n"
        exit 1
    fi

    # kill
    pid=`cat $agent_pid_file`
    if [ ! -z "$pid" ]; then
        kill $pid &>/dev/null
        sleep 1
    fi

    # rename
    mv $cov_start_file $cov_report_file

    # report
    convBin=${installPath}/codeCov/darwin/gocov
    convHtml=${installPath}/codeCov/darwin/gocov-html
    $convBin convert $cov_report_file | $convHtml > $cov_report_file_html
    printf "${info_msg}report:\n"
    printf "${info_msg}--> "
    tail -1 $agent_log_file | awk '{print $1,$2}'
    printf "${info_msg}--> detail: ${cov_report_url}\n"

    # clear
    rm -rf $agent_pid_file
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
    "report" )
        stop
        start
        ;;
    * )
        printf "unknow command!!! [usage: start、stop]\n"
        exit 1
        ;;
esac