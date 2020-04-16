#!/bin/bash

# To run this shell script to set $GOROOT; First run will wait a few minutes, keep patience!
#   `sh install.sh go1.13`
#   `export GOROOT=/tmp/recorder-go1.13`
#   `export PATH=${GOROOT}/bin:${PATH}`

# only support amd64
if [ `getconf LONG_BIT` != "64" ] ; then
    echo "only support amd64"
    exit
fi

function install(){
    VERSION=$1

    # base set
    GIT_URL="https://github.com/didichuxing/sharingan-go"

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
    echo "wget ${download_url} -O ${tmp_file}"
    wget ${download_url} -O ${tmp_file}

    # tar && set version
    tar -xzf ${tmp_file} -C ${recorder_dir} --strip-components=1

    # rm
    rm -rf ${tmp_file}
}

install $1
