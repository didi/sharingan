#!/bin/bash

# To run this shell script to set $GOROOT; First run will wait a few minutes, keep patience!
#   `sh install.sh go1.13`
#   `export GOROOT=/tmp/recorder-go1.13`
#   `export PATH=${GOROOT}/bin:${PATH}`

# set msg
info_msg="\033[;32m[INFO]\033[0m\t"
warn_msg="\033[;33m[WARN]\033[0m\t"
error_msg="\033[;31m[ERROR]\033[0m\t"

# only support amd64
if [ `getconf LONG_BIT` != "64" ] ; then
    printf "${warn_msg}-> only support amd64!!!\n"
    exit
fi

# nead version param
if [ ! -n "$1" ] ; then
    printf "${warn_msg}-> please input version!!!\n"
    printf "${info_msg}--> usage: sh install.sh go1.13\n"
    exit
fi

# only support go1.10 ~ go1.14
if [ "$1" != "go1.10" ] && [ "$1" != "go1.11" ] && [ "$1" != "go1.12" ] && [ "$1" != "go1.13" ] && [ "$1" != "go1.14" ]; then
    printf "${warn_msg}-> only support go1.10 ~ go1.14!!!\n"
    printf "${info_msg}--> usage: sh install.sh go1.13\n"
    exit
fi

function install(){
    VERSION=$1

    # base set
    GIT_URL="https://github.com/didi/sharingan-go"

    # param
    uname=`uname`
    uname=`echo $uname | tr '[:upper:]' '[:lower:]'`
    file_name=${VERSION}.${uname}-amd64
    recorder_dir=/tmp/recorder-${VERSION}
    should_update=true

    rm -rf ${recorder_dir} && mkdir -p ${recorder_dir}
    tmp_file=${recorder_dir}.tar.gz

    # download
    download_url=${GIT_URL}/releases/download/${VERSION}.recorder/${file_name}.tar.gz
    wget ${download_url} -O ${tmp_file}

    # tar && set version
    tar -xzf ${tmp_file} -C ${recorder_dir} --strip-components=1

    # rm
    rm -rf ${tmp_file}
}

install $1
