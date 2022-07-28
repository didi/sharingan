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

# set github about go
git_go='https://raw.githubusercontent.com/didi/sharingan-go/recorder'
install_go_md='https://github.com/golang/go#download-and-install'
VERSION="go1.13"

function install_go() {
        curl "$git_go/install/$VERSION" >> /dev/null
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at installing $VERSION : 'curl $git_go/install/$VERSION' \n"
            printf "${error_msg}Please install $VERSION manually (refer to $install_go_md) and try again!!!  \n"
            printf "${warn_msg}Please execute 'export GOROOT=/tmp/recorder-$VERSION && export PATH=\$GOROOT/bin:\$PATH' after installing $VERSION!!!\n"
            exit 1
        fi
        curl "$git_go/install/$VERSION" | sh
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at installing $VERSION : 'curl $git_go/install/$VERSION | sh' \n"
            printf "${error_msg}Please install $VERSION manually (refer to $install_go_md) and try again!!!  \n"
            printf "${warn_msg}Please execute 'export GOROOT=/tmp/recorder-$VERSION && export PATH=\$GOROOT/bin:\$PATH' after installing $VERSION!!!\n"
            exit 1
        fi
        export GOROOT="/tmp/recorder-$VERSION"
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at setting GOROOT, please check!!!\n"
            exit 1
        fi
        export PATH=$GOROOT/bin:$PATH
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at updating PATH, please check!!!\n"
            exit 1
        fi

        printf "${info_msg}GOROOT=$GOROOT \n"
}

# install go
function install_go_tar() {
        # get go url
        releases_go='https://github.com/didi/sharingan-go/releases/download/'
        if [[ "$OSTYPE" =~ ^darwin.* ]]; then
            goSuffix="$VERSION.recorder/$VERSION.darwin-amd64.tar.gz"
        elif [[ "$OSTYPE" =~ ^linux.* ]]; then
            goSuffix="$VERSION.recorder/$VERSION.linux-amd64.tar.gz"
        else
            printf "${error_msg} Unknown system type! build failed at installing go! \n"
            exit 1
        fi

        # install golang
        recorder_dir=/tmp/recorder-${VERSION}
        wget "$releases_go/$goSuffix" -O "$recorder_dir.tar.gz"
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at downloading $VERSION : wget $releases_go/$goSuffix -O $recorder_dir.tar.gz !!!\n"
            printf "${error_msg}Please install $VERSION manually (refer to $install_go_md) and try again!!!  \n"
            printf "${warn_msg}Please execute 'export GOROOT=/tmp/recorder-$VERSION && export PATH=\$GOROOT/bin:\$PATH' after installing $VERSION!!!\n"
            exit 1
        fi
        rm -rf $recorder_dir && mkdir -p $recorder_dir
        tar -xzf "$recorder_dir.tar.gz" -C ${recorder_dir} --strip-components=1
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at installing $VERSION, please check!!!\n"
            exit 1
        fi
        rm -rf "$recorder_dir.tar.gz"

        # export
        export GOROOT="/tmp/recorder-$VERSION"
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at setting GOROOT, please check!!!\n"
            exit 1
        fi
        export PATH=$GOROOT/bin:$PATH
        if [ $? -ne 0 ]; then
            printf "${error_msg}build failed at updating PATH, please check!!!\n"
            exit 1
        fi

        printf "${info_msg}GOROOT=$GOROOT \n"
}

# build
function build() {
    if [ -z $GOROOT  ];then
        install_go
    fi

    which go &> /dev/null
    if [ $? -ne 0 ]; then
        printf "${error_msg}build failed at executing 'which go', please update \$PATH!!!\n"
        exit 1
    fi
    goVer=`go version`
    if [[ $goVer < 'go version go1.11' ]];then
        printf "${warn_msg}current is $goVer, go mod requires at least go1.11!!!\n"
        install_go
    fi

    if [ -z $GOPATH  ];then
        mkdir -p /tmp/replayer
        export GOPATH=/tmp/replayer
        printf "${warn_msg}No GOPATH, so setting GOPATH=/tmp/replayer \n"
    fi

    cd $root
    go clean -modcache
    printf "${info_msg}go mod download, please wait~ \n"
    go mod download
    if [ $? -ne 0 ]; then
        printf "${error_msg}build failed at executing go mod download, please check \$GOPROXY or try again!!!\n"
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
